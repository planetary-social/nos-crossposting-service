//go:build wireinject
// +build wireinject

package di

import (
	"context"
	"database/sql"

	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain/notifications"
)

func BuildService(context.Context, config.Config) (Service, func(), error) {
	wire.Build(
		NewService,

		portsSet,
		applicationSet,
		sqliteAdaptersSet,
		//downloaderSet,
		//generatorSet,
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
		//downloaderSet,
		//generatorSet,
		pubsubSet,
		loggingSet,
		integrationAdaptersSet,
	)
	return IntegrationService{}, nil, nil
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

var downloaderSet = wire.NewSet(
	app.NewDownloader,
)

var generatorSet = wire.NewSet(
	notifications.NewGenerator,
)
