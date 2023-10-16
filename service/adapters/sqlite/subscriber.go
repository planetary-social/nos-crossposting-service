package sqlite

import (
	"context"
	"database/sql"

	watermillsql "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
)

type Subscriber struct {
	watermillSubscriber *watermillsql.Subscriber
	offsetsAdapter      watermillsql.OffsetsAdapter
	schema              SqliteSchema
	db                  *sql.DB
}

func NewSubscriber(
	watermillSubscriber *watermillsql.Subscriber,
	offsetsAdapter watermillsql.OffsetsAdapter,
	schema SqliteSchema,
	db *sql.DB,
) *Subscriber {
	return &Subscriber{
		watermillSubscriber: watermillSubscriber,
		offsetsAdapter:      offsetsAdapter,
		schema:              schema,
		db:                  db,
	}
}

func (s *Subscriber) SubscribeToTweetCreated(ctx context.Context) (<-chan *message.Message, error) {
	return s.watermillSubscriber.Subscribe(ctx, TweetCreatedTopic)
}

func (s *Subscriber) TweetCreatedQueueLength(ctx context.Context) (int, error) {
	offsetsQuery, offsetsQueryArgs := s.offsetsAdapter.NextOffsetQuery(TweetCreatedTopic, consumerGroupName)

	selectQuery := `
		SELECT COUNT(*)
		FROM ` + s.schema.MessagesTable(TweetCreatedTopic) + `
		WHERE offset > (` + offsetsQuery + `)`

	row := s.db.QueryRowContext(ctx, selectQuery, offsetsQueryArgs...)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "error calling row scan")
	}

	return count, nil
}
