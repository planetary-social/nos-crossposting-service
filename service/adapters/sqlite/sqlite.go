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

type TestAdapters struct {
	SessionRepository        *SessionRepository
	AccountRepository        *AccountRepository
	PublicKeyRepository      *PublicKeyRepository
	ProcessedEventRepository *ProcessedEventRepository
	UserTokensRepository     *UserTokensRepository
	Publisher                *Publisher
}

type TestedItems struct {
	TransactionProvider *TestTransactionProvider
	Migrations          *Migrations
	Subscriber          *Subscriber
	MigrationsStorage   *MigrationsStorage
	PubSub              *PubSub
}

func Open(conf config.Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", conf.DatabasePath())
	if err != nil {
		return nil, errors.Wrap(err, "error opening the database")
	}

	return db, nil
}

type AdaptersFactoryFn = GenericAdaptersFactoryFn[app.Adapters]
type TransactionProvider = GenericTransactionProvider[app.Adapters]

func NewTransactionProvider(db *sql.DB, fn AdaptersFactoryFn) *TransactionProvider {
	return &TransactionProvider{
		db: db,
		fn: fn,
	}
}

type TestAdaptersFactoryFn = GenericAdaptersFactoryFn[TestAdapters]
type TestTransactionProvider = GenericTransactionProvider[TestAdapters]

func NewTestTransactionProvider(db *sql.DB, fn TestAdaptersFactoryFn) *TestTransactionProvider {
	return &TestTransactionProvider{
		db: db,
		fn: fn,
	}
}

type GenericAdaptersFactoryFn[T any] func(*sql.DB, *sql.Tx) (T, error)

type GenericTransactionProvider[T any] struct {
	db *sql.DB
	fn GenericAdaptersFactoryFn[T]
}

func NewGenericTransactionProvider[T any](db *sql.DB, fn GenericAdaptersFactoryFn[T]) *GenericTransactionProvider[T] {
	return &GenericTransactionProvider[T]{
		db: db,
		fn: fn,
	}
}

func (t *GenericTransactionProvider[T]) Transact(ctx context.Context, f func(context.Context, T) error) error {
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
