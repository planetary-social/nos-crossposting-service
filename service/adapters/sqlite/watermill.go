package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	watermillsql "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/boreq/errors"
)

func NewWatermillPublisher(
	tx *sql.Tx,
	logger watermill.LoggerAdapter,
	schemaAdapter watermillsql.SchemaAdapter,
) (*watermillsql.Publisher, error) {
	config := watermillsql.PublisherConfig{
		SchemaAdapter:        schemaAdapter,
		AutoInitializeSchema: false,
	}

	return watermillsql.NewPublisher(tx, config, logger)
}

func NewWatermillSubscriber(
	db *sql.DB,
	logger watermill.LoggerAdapter,
	schemaAdapter watermillsql.SchemaAdapter,
	offsetsAdapter watermillsql.OffsetsAdapter,
) (*watermillsql.Subscriber, error) {
	config := watermillsql.SubscriberConfig{
		ConsumerGroup:    "main",
		PollInterval:     30 * time.Second,
		ResendInterval:   30 * time.Second,
		RetryInterval:    30 * time.Second,
		SchemaAdapter:    schemaAdapter,
		OffsetsAdapter:   offsetsAdapter,
		InitializeSchema: false,
	}

	return watermillsql.NewSubscriber(db, config, logger)
}

func NewWatermillSchemaAdapter() watermillsql.SchemaAdapter {
	return SqliteSchema{
		GenerateMessagesTableName: func(topic string) string {
			return fmt.Sprintf("watermill_%s", topic)
		},
	}
}

func NewWatermillOffsetsAdapter() watermillsql.OffsetsAdapter {
	return SqliteOffsetsAdapter{
		GenerateMessagesOffsetsTableName: func(topic string) string {
			return fmt.Sprintf("watermill_offsets_%s", topic)
		},
	}
}

type SqliteSchema struct {
	GenerateMessagesTableName func(topic string) string
	SubscribeBatchSize        int
}

func (s SqliteSchema) SchemaInitializingQueries(topic string) []string {
	createMessagesTable := strings.Join([]string{
		"CREATE TABLE IF NOT EXISTS " + s.MessagesTable(topic) + " (",
		"`offset` INTEGER NOT NULL PRIMARY KEY,",
		"`uuid` VARCHAR(36) NOT NULL,",
		"`created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,",
		"`payload` JSON DEFAULT NULL,",
		"`metadata` JSON DEFAULT NULL",
		");",
	}, "\n")

	return []string{createMessagesTable}
}

func (s SqliteSchema) InsertQuery(topic string, msgs message.Messages) (string, []interface{}, error) {
	insertQuery := fmt.Sprintf(
		`INSERT INTO %s (uuid, payload, metadata) VALUES %s`,
		s.MessagesTable(topic),
		strings.TrimRight(strings.Repeat(`(?,?,?),`, len(msgs)), ","),
	)

	args, err := defaultInsertArgs(msgs)
	if err != nil {
		return "", nil, err
	}

	return insertQuery, args, nil
}

func (s SqliteSchema) batchSize() int {
	if s.SubscribeBatchSize == 0 {
		return 100
	}

	return s.SubscribeBatchSize
}

func (s SqliteSchema) SelectQuery(topic string, consumerGroup string, offsetsAdapter watermillsql.OffsetsAdapter) (string, []interface{}) {
	nextOffsetQuery, nextOffsetArgs := offsetsAdapter.NextOffsetQuery(topic, consumerGroup)
	selectQuery := `
		SELECT offset, uuid, payload, metadata FROM ` + s.MessagesTable(topic) + `
		WHERE 
			offset > (` + nextOffsetQuery + `)
		ORDER BY 
			offset ASC
		LIMIT ` + fmt.Sprintf("%d", s.batchSize())

	return selectQuery, nextOffsetArgs
}

func (s SqliteSchema) UnmarshalMessage(row watermillsql.Scanner) (watermillsql.Row, error) {
	r := watermillsql.Row{}
	err := row.Scan(&r.Offset, &r.UUID, &r.Payload, &r.Metadata)
	if err != nil {
		return watermillsql.Row{}, errors.Wrap(err, "could not scan message row")
	}

	msg := message.NewMessage(string(r.UUID), r.Payload)

	if r.Metadata != nil {
		err = json.Unmarshal(r.Metadata, &msg.Metadata)
		if err != nil {
			return watermillsql.Row{}, errors.Wrap(err, "could not unmarshal metadata as JSON")
		}
	}

	r.Msg = msg

	return r, nil
}

func (s SqliteSchema) MessagesTable(topic string) string {
	if s.GenerateMessagesTableName != nil {
		return s.GenerateMessagesTableName(topic)
	}
	return fmt.Sprintf("`watermill_%s`", topic)
}

func (s SqliteSchema) SubscribeIsolationLevel() sql.IsolationLevel {
	return sql.LevelSerializable
}

type SqliteOffsetsAdapter struct {
	GenerateMessagesOffsetsTableName func(topic string) string
}

func (a SqliteOffsetsAdapter) SchemaInitializingQueries(topic string) []string {
	return []string{`
		CREATE TABLE IF NOT EXISTS ` + a.MessagesOffsetsTable(topic) + ` (
		consumer_group VARCHAR(255) NOT NULL,
		offset_acked BIGINT,
		offset_consumed BIGINT NOT NULL,
		PRIMARY KEY(consumer_group)
	)`}
}

func (a SqliteOffsetsAdapter) AckMessageQuery(topic string, row watermillsql.Row, consumerGroup string) (string, []interface{}) {
	ackQuery := `INSERT INTO ` + a.MessagesOffsetsTable(topic) + ` (offset_consumed, offset_acked, consumer_group)
		VALUES (?, ?, ?) ON CONFLICT(consumer_group) DO UPDATE SET offset_consumed=excluded.offset_consumed, offset_acked=excluded.offset_acked`
	return ackQuery, []interface{}{row.Offset, row.Offset, consumerGroup}
}

func (a SqliteOffsetsAdapter) NextOffsetQuery(topic, consumerGroup string) (string, []interface{}) {
	return `SELECT COALESCE(
				(SELECT offset_acked
				 FROM ` + a.MessagesOffsetsTable(topic) + `
				 WHERE consumer_group=?
				), 0)`,
		[]interface{}{consumerGroup}
}

func (a SqliteOffsetsAdapter) MessagesOffsetsTable(topic string) string {
	if a.GenerateMessagesOffsetsTableName != nil {
		return a.GenerateMessagesOffsetsTableName(topic)
	}
	return fmt.Sprintf("`watermill_offsets_%s`", topic)
}

func (a SqliteOffsetsAdapter) ConsumedMessageQuery(topic string, row watermillsql.Row, consumerGroup string, consumerULID []byte) (string, []interface{}) {
	// offset_consumed is not queried anywhere, it's used only to detect race conditions with NextOffsetQuery.
	ackQuery := `INSERT INTO ` + a.MessagesOffsetsTable(topic) + ` (offset_consumed, consumer_group)
		VALUES (?, ?) ON CONFLICT(consumer_group) DO UPDATE SET offset_consumed=excluded.offset_consumed`
	return ackQuery, []interface{}{row.Offset, consumerGroup}
}

func defaultInsertArgs(msgs message.Messages) ([]interface{}, error) {
	var args []interface{}
	for _, msg := range msgs {
		metadata, err := json.Marshal(msg.Metadata)
		if err != nil {
			return nil, errors.Wrapf(err, "could not marshal metadata into JSON for message %s", msg.UUID)
		}

		args = append(args, msg.UUID, []byte(msg.Payload), metadata)
	}

	return args, nil
}
