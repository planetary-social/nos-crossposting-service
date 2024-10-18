//go:build wireinject
// +build wireinject

package di

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/mocks"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/content"
)

func BuildService(context.Context, config.Config) (Service, func(), error) {
	wire.Build(
		NewService,

		portsSet,
		applicationSet,
		sqliteAdaptersSet,
		downloaderSet,
		vanishSubscriberSet,
		memoryPubsubSet,
		sqlitePubsubSet,
		loggingSet,
		adaptersSet,
		tweetGeneratorSet,
		migrationsAdaptersSet,
	)
	return Service{}, nil, nil
}

func BuildTestAdapters(context.Context, testing.TB) (sqlite.TestedItems, func(), error) {
	wire.Build(
		wire.Struct(new(sqlite.TestedItems), "*"),

		sqliteTestAdaptersSet,
		sqlitePubsubSet,
		loggingSet,
		newTestAdaptersConfig,
		migrationsAdaptersSet,
	)
	return sqlite.TestedItems{}, nil, nil
}

type TestApplication struct {
	SendTweetHandler *app.SendTweetHandler

	CurrentTimeProvider  *mocks.CurrentTimeProvider
	UserTokensRepository *mocks.UserTokensRepository
	Twitter              *mocks.Twitter
}

func BuildTestApplication(tb testing.TB) (TestApplication, error) {
	wire.Build(
		wire.Struct(new(TestApplication), "*"),

		applicationSet,
		testAdaptersSet,
		mockTxAdaptersSet,

		fixtures.TestLogger,
	)
	return TestApplication{}, nil
}

func newTestAdaptersConfig(tb testing.TB) (config.Config, error) {
	return config.NewConfig(
		fixtures.SomeString(),
		fixtures.SomeString(),
		config.EnvironmentDevelopment,
		logging.LevelDebug,
		fixtures.SomeString(),
		fixtures.SomeString(),
		fixtures.SomeFile(tb),
		fixtures.SomeString(),
	)
}

type buildTransactionSqliteAdaptersDependencies struct {
	Logger logging.Logger
}

func buildTransactionSqliteAdapters(*sql.DB, *sql.Tx, buildTransactionSqliteAdaptersDependencies) (app.Adapters, error) {
	wire.Build(
		wire.Struct(new(app.Adapters), "*"),
		wire.FieldsOf(new(buildTransactionSqliteAdaptersDependencies), "Logger"),

		sqliteTxAdaptersSet,
		sqliteTxPubsubSet,
		sqlitePubsubSet,
	)
	return app.Adapters{}, nil
}

func buildTestTransactionSqliteAdapters(*sql.DB, *sql.Tx, buildTransactionSqliteAdaptersDependencies) (sqlite.TestAdapters, error) {
	wire.Build(
		wire.Struct(new(sqlite.TestAdapters), "*"),
		wire.FieldsOf(new(buildTransactionSqliteAdaptersDependencies), "Logger"),

		sqliteTxAdaptersSet,
		sqliteTxPubsubSet,
		sqlitePubsubSet,
	)
	return sqlite.TestAdapters{}, nil
}

var downloaderSet = wire.NewSet(
	app.NewDownloader,
)

var vanishSubscriberSet = wire.NewSet(
	app.NewVanishSubscriber,
)

var tweetGeneratorSet = wire.NewSet(
	content.NewTransformer,
	domain.NewTweetGenerator,
	wire.Bind(new(app.TweetGenerator), new(*domain.TweetGenerator)),
)
