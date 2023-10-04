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

	return nil
}
