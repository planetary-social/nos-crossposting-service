package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/migrations"
	"github.com/planetary-social/nos-crossposting-service/service/adapters"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
)

var migrationsAdaptersSet = wire.NewSet(
	sqlite.NewMigrations,
	sqlite.NewMigrationFns,
	migrations.NewRunner,

	sqlite.NewMigrationsStorage,
	wire.Bind(new(migrations.Storage), new(*sqlite.MigrationsStorage)),

	adapters.NewLoggingMigrationsProgressCallback,
	wire.Bind(new(migrations.ProgressCallback), new(*adapters.LoggingMigrationsProgressCallback)),
)
