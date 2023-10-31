package sqlite

import (
	"database/sql"
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/migrations"
)

type MigrationsStorage struct {
	db *sql.DB
}

func NewMigrationsStorage(db *sql.DB) (*MigrationsStorage, error) {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations_state (
			name TEXT PRIMARY KEY,
			state TEXT
		);`,
	); err != nil {
		return nil, errors.Wrap(err, "error creating the state table")
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations_status (
			name TEXT PRIMARY KEY,
			status TEXT
		);`,
	); err != nil {
		return nil, errors.Wrap(err, "error creating the status table")
	}

	return &MigrationsStorage{db: db}, nil
}

func (b *MigrationsStorage) LoadState(name string) (migrations.State, error) {
	var marshaledState string

	row := b.db.QueryRow("SELECT state FROM migrations_state WHERE name=$1", name)
	if err := row.Scan(&marshaledState); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, migrations.ErrStateNotFound
		}
		return nil, errors.Wrap(err, "error querying the database")
	}

	var state migrations.State
	if err := json.Unmarshal([]byte(marshaledState), &state); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling state")
	}

	return state, nil
}

func (b *MigrationsStorage) SaveState(name string, state migrations.State) error {
	marshaledState, err := json.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "error marshaling state")
	}

	if _, err := b.db.Exec(`
INSERT INTO migrations_state(name, state)
VALUES ($1, $2)
ON CONFLICT(name) DO UPDATE SET
  state=excluded.state`,
		name, marshaledState); err != nil {
		return errors.Wrap(err, "error running the query")
	}

	return nil
}

func (b *MigrationsStorage) LoadStatus(name string) (migrations.Status, error) {
	var marshaledStatus string

	row := b.db.QueryRow("SELECT status FROM migrations_status WHERE name=$1", name)
	if err := row.Scan(&marshaledStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return migrations.Status{}, migrations.ErrStatusNotFound
		}
		return migrations.Status{}, errors.Wrap(err, "error querying the database")
	}

	status, err := unmarshalStatus(marshaledStatus)
	if err != nil {
		return migrations.Status{}, errors.Wrap(err, "error unmarshaling status")
	}

	return status, nil
}

func (b *MigrationsStorage) SaveStatus(name string, status migrations.Status) error {
	marshaledStatus, err := marshalStatus(status)
	if err != nil {
		return errors.Wrap(err, "error marshaling status")
	}

	if _, err := b.db.Exec(`
INSERT INTO migrations_status(name, status)
VALUES ($1, $2)
ON CONFLICT(name) DO UPDATE SET
  status=excluded.status`,
		name, marshaledStatus); err != nil {
		return errors.Wrap(err, "error running the query")
	}

	return nil
}

const (
	statusFailed   = "failed"
	statusFinished = "finished"
)

func marshalStatus(status migrations.Status) (string, error) {
	switch status {
	case migrations.StatusFailed:
		return statusFailed, nil
	case migrations.StatusFinished:
		return statusFinished, nil
	default:
		return "", errors.New("unknown status")
	}
}

func unmarshalStatus(status string) (migrations.Status, error) {
	switch status {
	case statusFailed:
		return migrations.StatusFailed, nil
	case statusFinished:
		return migrations.StatusFinished, nil
	default:
		return migrations.Status{}, errors.New("unknown status")
	}
}
