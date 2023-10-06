package sqlite

import (
	"context"
	"database/sql"

	"github.com/boreq/errors"
)

type Migrations struct {
	db *sql.DB
}

func NewMigrations(db *sql.DB) *Migrations {
	return &Migrations{
		db: db,
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

	return nil
}
