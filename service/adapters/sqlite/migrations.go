package sqlite

import (
	"context"
	"database/sql"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/migrations"
)

func NewMigrations(fns *MigrationFns) (migrations.Migrations, error) {
	return migrations.NewMigrations([]migrations.Migration{
		migrations.MustNewMigration("initial", fns.Initial),
		migrations.MustNewMigration("create_pubsub_tables", fns.CreatePubsubTables),
	})
}

type MigrationFns struct {
	db     *sql.DB
	pubsub *PubSub
}

func NewMigrationFns(db *sql.DB, pubsub *PubSub) *MigrationFns {
	return &MigrationFns{db: db, pubsub: pubsub}
}

func (m *MigrationFns) Initial(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
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

	return nil
}

func (m *MigrationFns) CreatePubsubTables(ctx context.Context, state migrations.State, saveStateFunc migrations.SaveStateFunc) error {
	for _, query := range m.pubsub.InitializingQueries() {
		if _, err := m.db.Exec(query); err != nil {
			return errors.Wrapf(err, "error initializing pubsub")
		}
	}

	return nil
}
