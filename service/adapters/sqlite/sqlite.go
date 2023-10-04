package sqlite

import (
	"context"
	"database/sql"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	_ "github.com/mattn/go-sqlite3"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
)

func Open(config config.Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", config.DatabasePath())
	if err != nil {
		return nil, errors.Wrap(err, "error opening the database")
	}

	return db, nil
}

type AdaptersFactoryFn func(*sql.DB, *sql.Tx) (app.Adapters, error)

type TransactionProvider struct {
	db *sql.DB
	fn AdaptersFactoryFn
}

func NewTransactionProvider(db *sql.DB, fn AdaptersFactoryFn) *TransactionProvider {
	return &TransactionProvider{
		db: db,
		fn: fn,
	}
}

func (t *TransactionProvider) Transact(ctx context.Context, f func(context.Context, app.Adapters) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "error starting the transaction")
	}

	adapters, err := t.fn(t.db, tx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = multierror.Append(err, errors.Wrap(rollbackErr, "rollback error"))
		}
		return errors.Wrap(err, "error building the adapters")
	}

	if err := f(ctx, adapters); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = multierror.Append(err, errors.Wrap(rollbackErr, "rollback error"))
		}
		return errors.Wrap(err, "error calling the provided function")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "error committing the transaction")
	}

	return nil
}
