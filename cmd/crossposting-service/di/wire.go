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
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
)

func BuildService(context.Context, config.Config) (Service, func(), error) {
	wire.Build(
		NewService,

		portsSet,
		applicationSet,
		sqliteAdaptersSet,
		downloaderSet,
		pubsubSet,
		loggingSet,
		adaptersSet,
	)
	return Service{}, nil, nil
}

type IntegrationService struct {
	Service Service
}

func BuildIntegrationService(context.Context, config.Config) (IntegrationService, func(), error) {
	wire.Build(
		wire.Struct(new(IntegrationService), "*"),

		NewService,

		portsSet,
		applicationSet,
		sqliteAdaptersSet,
		downloaderSet,
		//generatorSet,
		pubsubSet,
		loggingSet,
		integrationAdaptersSet,
	)
	return IntegrationService{}, nil, nil
}

func BuildTestAdapters(context.Context, testing.TB) (sqlite.TestedItems, func(), error) {
	wire.Build(
		wire.Struct(new(sqlite.TestedItems), "*"),

		sqliteTestAdaptersSet,
		loggingSet,
		newTestAdaptersConfig,
	)
	return sqlite.TestedItems{}, nil, nil
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
	)
}

type buildTransactionSqliteAdaptersDependencies struct {
	//LoggerAdapter watermill.LoggerAdapter
}

func buildTransactionSqliteAdapters(*sql.DB, *sql.Tx, buildTransactionSqliteAdaptersDependencies) (app.Adapters, error) {
	wire.Build(
		wire.Struct(new(app.Adapters), "*"),
		//wire.FieldsOf(new(buildTransactionSqliteAdaptersDependencies), "LoggerAdapter"),

		sqliteTxAdaptersSet,
	)
	return app.Adapters{}, nil
}

func buildTestTransactionSqliteAdapters(*sql.DB, *sql.Tx, buildTransactionSqliteAdaptersDependencies) (sqlite.TestAdapters, error) {
	wire.Build(
		wire.Struct(new(sqlite.TestAdapters), "*"),
		//wire.FieldsOf(new(buildTransactionSqliteAdaptersDependencies), "LoggerAdapter"),

		sqliteTxAdaptersSet,
	)
	return sqlite.TestAdapters{}, nil
}

var downloaderSet = wire.NewSet(
	app.NewDownloader,
)
