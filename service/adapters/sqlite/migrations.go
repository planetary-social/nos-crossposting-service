package sqlite

import (
	"context"
	"database/sql"

	watermillsql "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
	"github.com/boreq/errors"
)

type Migrations struct {
	db                     *sql.DB
	watermillSchemaAdapter watermillsql.SchemaAdapter
	watermilOffsetsAdapter watermillsql.OffsetsAdapter
	pubsub                 *PubSub
}

func NewMigrations(
	db *sql.DB,
	watermillSchemaAdapter watermillsql.SchemaAdapter,
	watermillOffsetsAdapter watermillsql.OffsetsAdapter,
	pubsub *PubSub,
) *Migrations {
	return &Migrations{
		db:                     db,
		watermillSchemaAdapter: watermillSchemaAdapter,
		watermilOffsetsAdapter: watermillOffsetsAdapter,
		pubsub:                 pubsub,
	}
}

func (m *Migrations) Execute(ctx context.Context) error {
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			account_id TEXT PRIMARY KEY,
			twitter_id INTEGER UNIQUE
		);`,
	)
	if err != nil {
		return errors.Wrap(err, "error creating the accounts table")
	}

	_, err = m.db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT PRIMARY KEY,
			account_id TEXT,
			created_at INTEGER,
			FOREIGN KEY(account_id) REFERENCES accounts(account_id)
		);`,
	)
	if err != nil {
		return errors.Wrap(err, "error creating the sessions table")
	}

	_, err = m.db.Exec(`
		CREATE TABLE IF NOT EXISTS public_keys (
			account_id TEXT,
			public_key TEXT,
			created_at INTEGER,
			PRIMARY KEY(account_id, public_key),
			FOREIGN KEY(account_id) REFERENCES accounts(account_id)
		);`,
	)
	if err != nil {
		return errors.Wrap(err, "error creating the public keys table")
	}

	_, err = m.db.Exec(`
		CREATE TABLE IF NOT EXISTS processed_events (
			twitter_id INTEGER,
			event_id TEXT,
			PRIMARY KEY(twitter_id, event_id)
		);`,
	)
	if err != nil {
		return errors.Wrap(err, "error creating the processed events table")
	}

	_, err = m.db.Exec(`
		CREATE TABLE IF NOT EXISTS user_tokens (
			account_id TEXT PRIMARY KEY,
			access_token TEXT,
			access_secret TEXT,
			FOREIGN KEY(account_id) REFERENCES accounts(account_id)
		);`,
	)
	if err != nil {
		return errors.Wrap(err, "error creating the user tokens table")
	}

	for _, topic := range []string{TweetCreatedTopic} {
		for _, query := range m.watermillSchemaAdapter.SchemaInitializingQueries(topic) {
			if _, err = m.db.Exec(query); err != nil {
				return errors.Wrapf(err, "error initializing watermill schema for topic '%s'", topic)
			}
		}

		for _, query := range m.watermilOffsetsAdapter.SchemaInitializingQueries(topic) {
			if _, err = m.db.Exec(query); err != nil {
				return errors.Wrapf(err, "error initializing watermill offsets for topic '%s'", topic)
			}
		}
	}

	for _, query := range m.pubsub.InitializingQueries() {
		if _, err = m.db.Exec(query); err != nil {
			return errors.Wrapf(err, "error initializing pubsub")
		}
	}

	return nil
}
