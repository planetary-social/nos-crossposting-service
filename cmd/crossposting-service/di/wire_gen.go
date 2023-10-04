// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"context"
	"database/sql"

	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/service/adapters"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/prometheus"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/pubsub"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain/notifications"
	"github.com/planetary-social/nos-crossposting-service/service/ports/http"
)

// Injectors from wire.go:

func BuildService(contextContext context.Context, configConfig config.Config) (Service, func(), error) {
	memoryEventWasAlreadySavedCache := adapters.NewMemoryEventWasAlreadySavedCache()
	logger, err := newLogger(configConfig)
	if err != nil {
		return Service{}, nil, err
	}
	db, cleanup, err := newSqliteDB(configConfig, logger)
	if err != nil {
		return Service{}, nil, err
	}
	diBuildTransactionSqliteAdaptersDependencies := buildTransactionSqliteAdaptersDependencies{}
	adaptersFactoryFn := newAdaptersFactoryFn(diBuildTransactionSqliteAdaptersDependencies)
	transactionProvider := sqlite.NewTransactionProvider(db, adaptersFactoryFn)
	prometheusPrometheus, err := prometheus.NewPrometheus(logger)
	if err != nil {
		cleanup()
		return Service{}, nil, err
	}
	saveReceivedEventHandler := app.NewSaveReceivedEventHandler(memoryEventWasAlreadySavedCache, transactionProvider, logger, prometheusPrometheus)
	getRelaysHandler := app.NewGetRelaysHandler(transactionProvider, prometheusPrometheus)
	getPublicKeysHandler := app.NewGetPublicKeysHandler(transactionProvider, prometheusPrometheus)
	getTokensHandler := app.NewGetTokensHandler(transactionProvider, prometheusPrometheus)
	receivedEventPubSub := pubsub.NewReceivedEventPubSub()
	getEventsHandler := app.NewGetEventsHandler(transactionProvider, receivedEventPubSub, prometheusPrometheus)
	getNotificationsHandler := app.NewGetNotificationsHandler(transactionProvider, prometheusPrometheus)
	getSessionAccountHandler := app.NewGetSessionAccountHandler(transactionProvider, logger, prometheusPrometheus)
	idGenerator := adapters.NewIDGenerator()
	loginOrRegisterHandler := app.NewLoginOrRegisterHandler(transactionProvider, idGenerator, idGenerator, logger, prometheusPrometheus)
	application := app.Application{
		SaveReceivedEvent: saveReceivedEventHandler,
		GetRelays:         getRelaysHandler,
		GetPublicKeys:     getPublicKeysHandler,
		GetTokens:         getTokensHandler,
		GetEvents:         getEventsHandler,
		GetNotifications:  getNotificationsHandler,
		GetSessionAccount: getSessionAccountHandler,
		LoginOrRegister:   loginOrRegisterHandler,
	}
	server := http.NewServer(configConfig, application, logger)
	metricsServer := http.NewMetricsServer(prometheusPrometheus, configConfig, logger)
	migrations := sqlite.NewMigrations(db)
	service := NewService(application, server, metricsServer, memoryEventWasAlreadySavedCache, migrations)
	return service, func() {
		cleanup()
	}, nil
}

func BuildIntegrationService(contextContext context.Context, configConfig config.Config) (IntegrationService, func(), error) {
	memoryEventWasAlreadySavedCache := adapters.NewMemoryEventWasAlreadySavedCache()
	logger, err := newLogger(configConfig)
	if err != nil {
		return IntegrationService{}, nil, err
	}
	db, cleanup, err := newSqliteDB(configConfig, logger)
	if err != nil {
		return IntegrationService{}, nil, err
	}
	diBuildTransactionSqliteAdaptersDependencies := buildTransactionSqliteAdaptersDependencies{}
	adaptersFactoryFn := newAdaptersFactoryFn(diBuildTransactionSqliteAdaptersDependencies)
	transactionProvider := sqlite.NewTransactionProvider(db, adaptersFactoryFn)
	prometheusPrometheus, err := prometheus.NewPrometheus(logger)
	if err != nil {
		cleanup()
		return IntegrationService{}, nil, err
	}
	saveReceivedEventHandler := app.NewSaveReceivedEventHandler(memoryEventWasAlreadySavedCache, transactionProvider, logger, prometheusPrometheus)
	getRelaysHandler := app.NewGetRelaysHandler(transactionProvider, prometheusPrometheus)
	getPublicKeysHandler := app.NewGetPublicKeysHandler(transactionProvider, prometheusPrometheus)
	getTokensHandler := app.NewGetTokensHandler(transactionProvider, prometheusPrometheus)
	receivedEventPubSub := pubsub.NewReceivedEventPubSub()
	getEventsHandler := app.NewGetEventsHandler(transactionProvider, receivedEventPubSub, prometheusPrometheus)
	getNotificationsHandler := app.NewGetNotificationsHandler(transactionProvider, prometheusPrometheus)
	getSessionAccountHandler := app.NewGetSessionAccountHandler(transactionProvider, logger, prometheusPrometheus)
	idGenerator := adapters.NewIDGenerator()
	loginOrRegisterHandler := app.NewLoginOrRegisterHandler(transactionProvider, idGenerator, idGenerator, logger, prometheusPrometheus)
	application := app.Application{
		SaveReceivedEvent: saveReceivedEventHandler,
		GetRelays:         getRelaysHandler,
		GetPublicKeys:     getPublicKeysHandler,
		GetTokens:         getTokensHandler,
		GetEvents:         getEventsHandler,
		GetNotifications:  getNotificationsHandler,
		GetSessionAccount: getSessionAccountHandler,
		LoginOrRegister:   loginOrRegisterHandler,
	}
	server := http.NewServer(configConfig, application, logger)
	metricsServer := http.NewMetricsServer(prometheusPrometheus, configConfig, logger)
	migrations := sqlite.NewMigrations(db)
	service := NewService(application, server, metricsServer, memoryEventWasAlreadySavedCache, migrations)
	integrationService := IntegrationService{
		Service: service,
	}
	return integrationService, func() {
		cleanup()
	}, nil
}

func buildTransactionSqliteAdapters(db *sql.DB, tx *sql.Tx, diBuildTransactionSqliteAdaptersDependencies buildTransactionSqliteAdaptersDependencies) (app.Adapters, error) {
	accountRepository, err := sqlite.NewAccountRepository(tx)
	if err != nil {
		return app.Adapters{}, err
	}
	sessionRepository, err := sqlite.NewSessionRepository(tx)
	if err != nil {
		return app.Adapters{}, err
	}
	appAdapters := app.Adapters{
		Accounts: accountRepository,
		Sessions: sessionRepository,
	}
	return appAdapters, nil
}

// wire.go:

type IntegrationService struct {
	Service Service
}

type buildTransactionSqliteAdaptersDependencies struct {
}

var downloaderSet = wire.NewSet(app.NewDownloader)

var generatorSet = wire.NewSet(notifications.NewGenerator)
