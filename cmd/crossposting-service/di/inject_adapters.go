package di

import (
	"context"
	"database/sql"

	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/adapters"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/mocks"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/prometheus"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/twitter"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

var sqliteAdaptersSet = wire.NewSet(
	newSqliteDB,

	sqlite.NewTransactionProvider,
	wire.Bind(new(app.TransactionProvider), new(*sqlite.TransactionProvider)),

	newAdaptersFactoryFn,

	wire.Struct(new(buildTransactionSqliteAdaptersDependencies), "*"),
)

var sqliteTestAdaptersSet = wire.NewSet(
	newSqliteDB,

	sqlite.NewTestTransactionProvider,

	newTestAdaptersFactoryFn,

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
	newPurplePages,
	wire.Bind(new(app.RelaySource), new(*adapters.RelaySource)),

	adapters.NewRelayEventDownloader,
	wire.Bind(new(app.RelayEventDownloader), new(*adapters.RelayEventDownloader)),

	twitter.NewTwitter,
	twitter.NewDevelopmentTwitter,
	selectTwitterAdapterDependingOnConfig,

	adapters.NewTwitterAccountDetailsCache,
	wire.Bind(new(app.TwitterAccountDetailsCache), new(*adapters.TwitterAccountDetailsCache)),

	adapters.NewCurrentTimeProvider,
	wire.Bind(new(app.CurrentTimeProvider), new(*adapters.CurrentTimeProvider)),
)

var testAdaptersSet = wire.NewSet(
	prometheus.NewPrometheus,
	wire.Bind(new(app.Metrics), new(*prometheus.Prometheus)),

	mocks.NewTwitter,
	wire.Bind(new(app.Twitter), new(*mocks.Twitter)),

	mocks.NewCurrentTimeProvider,
	wire.Bind(new(app.CurrentTimeProvider), new(*mocks.CurrentTimeProvider)),
)

var mockTxAdaptersSet = wire.NewSet(
	mocks.NewTransactionProvider,
	wire.Bind(new(app.TransactionProvider), new(*mocks.TransactionProvider)),

	wire.Struct(new(app.Adapters), "*"),

	mocks.NewAccountRepository,
	wire.Bind(new(app.AccountRepository), new(*mocks.AccountRepository)),

	mocks.NewSessionRepository,
	wire.Bind(new(app.SessionRepository), new(*mocks.SessionRepository)),

	mocks.NewPublicKeyRepository,
	wire.Bind(new(app.PublicKeyRepository), new(*mocks.PublicKeyRepository)),

	mocks.NewProcessedEventRepository,
	wire.Bind(new(app.ProcessedEventRepository), new(*mocks.ProcessedEventRepository)),

	mocks.NewUserTokensRepository,
	wire.Bind(new(app.UserTokensRepository), new(*mocks.UserTokensRepository)),

	mocks.NewPublisher,
	wire.Bind(new(app.Publisher), new(*mocks.Publisher)),
)

var purplePagesAddresses = []domain.RelayAddress{
	domain.MustNewRelayAddress("wss://purplepag.es"),
	domain.MustNewRelayAddress("wss://relay.nos.social"),
}

func newPurplePages(ctx context.Context, logger logging.Logger, metrics app.Metrics) ([]*adapters.CachedPurplePages, error) {
	var result []*adapters.CachedPurplePages

	for _, address := range purplePagesAddresses {
		v, err := adapters.NewPurplePages(ctx, address, logger, metrics)
		if err != nil {
			return nil, errors.Wrap(err, "error creating purple pages")
		}

		result = append(result, adapters.NewCachedPurplePages(logger, v))
	}

	return result, nil
}

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

func newSqliteDB(conf config.Config, logger logging.Logger) (*sql.DB, func(), error) {
	v, err := sqlite.Open(conf)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error opening sqlite")
	}

	return v, func() {
		if err := v.Close(); err != nil {
			logger.Error().WithError(err).Message("error closing sqlite")
		}
	}, nil
}

func selectTwitterAdapterDependingOnConfig(
	conf config.Config,
	productionAdapter *twitter.Twitter,
	developmentAdapter *twitter.DevelopmentTwitter,
) app.Twitter {
	if conf.Environment() == config.EnvironmentDevelopment {
		return developmentAdapter
	}
	return productionAdapter
}
