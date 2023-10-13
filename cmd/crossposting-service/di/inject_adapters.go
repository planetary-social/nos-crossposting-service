package di

import (
	"database/sql"

	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/adapters"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/prometheus"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/twitter"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
)

var sqliteAdaptersSet = wire.NewSet(
	newSqliteDB,

	sqlite.NewTransactionProvider,
	wire.Bind(new(app.TransactionProvider), new(*sqlite.TransactionProvider)),

	newAdaptersFactoryFn,

	sqlite.NewMigrations,

	wire.Struct(new(buildTransactionSqliteAdaptersDependencies), "*"),
)

var sqliteTestAdaptersSet = wire.NewSet(
	newSqliteDB,

	sqlite.NewTestTransactionProvider,

	newTestAdaptersFactoryFn,

	sqlite.NewMigrations,

	wire.Struct(new(buildTransactionSqliteAdaptersDependencies), "*"),
)

var sqliteTxAdaptersSet = wire.NewSet(
	sqlite.NewAccountRepository,
	wire.Bind(new(app.AccountRepository), new(*sqlite.AccountRepository)),

	sqlite.NewSessionRepository,
	wire.Bind(new(app.SessionRepository), new(*sqlite.SessionRepository)),

	sqlite.NewPublicKeyRepository,
	wire.Bind(new(app.PublicKeyRepository), new(*sqlite.PublicKeyRepository)),

	sqlite.NewProcessedEventRepository,
	wire.Bind(new(app.ProcessedEventRepository), new(*sqlite.ProcessedEventRepository)),

	sqlite.NewUserTokensRepository,
	wire.Bind(new(app.UserTokensRepository), new(*sqlite.UserTokensRepository)),
)

var adaptersSet = wire.NewSet(
	prometheus.NewPrometheus,
	wire.Bind(new(app.Metrics), new(*prometheus.Prometheus)),

	adapters.NewIDGenerator,
	wire.Bind(new(app.SessionIDGenerator), new(*adapters.IDGenerator)),
	wire.Bind(new(app.AccountIDGenerator), new(*adapters.IDGenerator)),

	adapters.NewRelaySource,
	adapters.NewPurplePages,
	wire.Bind(new(app.RelaySource), new(*adapters.RelaySource)),

	adapters.NewRelayEventDownloader,
	wire.Bind(new(app.RelayEventDownloader), new(*adapters.RelayEventDownloader)),

	twitter.NewTwitter,
	wire.Bind(new(app.Twitter), new(*twitter.Twitter)),
)

var integrationAdaptersSet = wire.NewSet(
	prometheus.NewPrometheus,
	wire.Bind(new(app.Metrics), new(*prometheus.Prometheus)),

	adapters.NewIDGenerator,
	wire.Bind(new(app.SessionIDGenerator), new(*adapters.IDGenerator)),
	wire.Bind(new(app.AccountIDGenerator), new(*adapters.IDGenerator)),

	adapters.NewPurplePages,
	wire.Bind(new(app.RelaySource), new(*adapters.PurplePages)),

	adapters.NewRelayEventDownloader,
	wire.Bind(new(app.RelayEventDownloader), new(*adapters.RelayEventDownloader)),

	twitter.NewTwitterMock,
	wire.Bind(new(app.Twitter), new(*twitter.TwitterMock)),
)

func newAdaptersFactoryFn(deps buildTransactionSqliteAdaptersDependencies) sqlite.AdaptersFactoryFn {
	return func(db *sql.DB, tx *sql.Tx) (app.Adapters, error) {
		return buildTransactionSqliteAdapters(db, tx, deps)
	}
}

func newTestAdaptersFactoryFn(deps buildTransactionSqliteAdaptersDependencies) sqlite.TestAdaptersFactoryFn {
	return func(db *sql.DB, tx *sql.Tx) (sqlite.TestAdapters, error) {
		return buildTestTransactionSqliteAdapters(db, tx, deps)
	}
}

func newSqliteDB(config config.Config, logger logging.Logger) (*sql.DB, func(), error) {
	v, err := sqlite.Open(config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error opening the sqlite database")
	}

	return v, func() {
		if err := v.Close(); err != nil {
			logger.Error().WithError(err).Message("error closing firestore")
		}
	}, nil
}
