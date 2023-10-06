// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/adapters"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/prometheus"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/pubsub"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/ports/http"
	"github.com/planetary-social/nos-crossposting-service/service/ports/memorypubsub"
)

// Injectors from wire.go:

func BuildService(contextContext context.Context, configConfig config.Config) (Service, func(), error) {
	logger, err := newLogger(configConfig)
	if err != nil {
		return Service{}, nil, err
	}
	db, cleanup, err := newSqliteDB(configConfig, logger)
	if err != nil {
		return Service{}, nil, err
	}
	diBuildTransactionSqliteAdaptersDependencies := buildTransactionSqliteAdaptersDependencies{}
	genericAdaptersFactoryFn := newAdaptersFactoryFn(diBuildTransactionSqliteAdaptersDependencies)
	genericTransactionProvider := sqlite.NewTransactionProvider(db, genericAdaptersFactoryFn)
	prometheusPrometheus, err := prometheus.NewPrometheus(logger)
	if err != nil {
		cleanup()
		return Service{}, nil, err
	}
	getSessionAccountHandler := app.NewGetSessionAccountHandler(genericTransactionProvider, logger, prometheusPrometheus)
	idGenerator := adapters.NewIDGenerator()
	loginOrRegisterHandler := app.NewLoginOrRegisterHandler(genericTransactionProvider, idGenerator, idGenerator, logger, prometheusPrometheus)
	linkPublicKeyHandler := app.NewLinkPublicKeyHandler(genericTransactionProvider, logger, prometheusPrometheus)
	application := app.Application{
		GetSessionAccount: getSessionAccountHandler,
		LoginOrRegister:   loginOrRegisterHandler,
		LinkPublicKey:     linkPublicKeyHandler,
	}
	server := http.NewServer(configConfig, application, logger)
	metricsServer := http.NewMetricsServer(prometheusPrometheus, configConfig, logger)
	receivedEventPubSub := pubsub.NewReceivedEventPubSub()
	purplePages, err := adapters.NewPurplePages(contextContext, logger)
	if err != nil {
		cleanup()
		return Service{}, nil, err
	}
	relaySource := adapters.NewRelaySource(logger, purplePages)
	relayEventDownloader := adapters.NewRelayEventDownloader(contextContext, logger, prometheusPrometheus)
	downloader := app.NewDownloader(genericTransactionProvider, receivedEventPubSub, logger, prometheusPrometheus, relaySource, relayEventDownloader)
	saveReceivedEventHandler := app.NewSaveReceivedEventHandler(genericTransactionProvider, logger, prometheusPrometheus)
	receivedEventSubscriber := memorypubsub.NewReceivedEventSubscriber(receivedEventPubSub, saveReceivedEventHandler, logger)
	migrations := sqlite.NewMigrations(db)
	service := NewService(application, server, metricsServer, downloader, receivedEventSubscriber, migrations)
	return service, func() {
		cleanup()
	}, nil
}

func BuildIntegrationService(contextContext context.Context, configConfig config.Config) (IntegrationService, func(), error) {
	logger, err := newLogger(configConfig)
	if err != nil {
		return IntegrationService{}, nil, err
	}
	db, cleanup, err := newSqliteDB(configConfig, logger)
	if err != nil {
		return IntegrationService{}, nil, err
	}
	diBuildTransactionSqliteAdaptersDependencies := buildTransactionSqliteAdaptersDependencies{}
	genericAdaptersFactoryFn := newAdaptersFactoryFn(diBuildTransactionSqliteAdaptersDependencies)
	genericTransactionProvider := sqlite.NewTransactionProvider(db, genericAdaptersFactoryFn)
	prometheusPrometheus, err := prometheus.NewPrometheus(logger)
	if err != nil {
		cleanup()
		return IntegrationService{}, nil, err
	}
	getSessionAccountHandler := app.NewGetSessionAccountHandler(genericTransactionProvider, logger, prometheusPrometheus)
	idGenerator := adapters.NewIDGenerator()
	loginOrRegisterHandler := app.NewLoginOrRegisterHandler(genericTransactionProvider, idGenerator, idGenerator, logger, prometheusPrometheus)
	linkPublicKeyHandler := app.NewLinkPublicKeyHandler(genericTransactionProvider, logger, prometheusPrometheus)
	application := app.Application{
		GetSessionAccount: getSessionAccountHandler,
		LoginOrRegister:   loginOrRegisterHandler,
		LinkPublicKey:     linkPublicKeyHandler,
	}
	server := http.NewServer(configConfig, application, logger)
	metricsServer := http.NewMetricsServer(prometheusPrometheus, configConfig, logger)
	receivedEventPubSub := pubsub.NewReceivedEventPubSub()
	purplePages, err := adapters.NewPurplePages(contextContext, logger)
	if err != nil {
		cleanup()
		return IntegrationService{}, nil, err
	}
	relayEventDownloader := adapters.NewRelayEventDownloader(contextContext, logger, prometheusPrometheus)
	downloader := app.NewDownloader(genericTransactionProvider, receivedEventPubSub, logger, prometheusPrometheus, purplePages, relayEventDownloader)
	saveReceivedEventHandler := app.NewSaveReceivedEventHandler(genericTransactionProvider, logger, prometheusPrometheus)
	receivedEventSubscriber := memorypubsub.NewReceivedEventSubscriber(receivedEventPubSub, saveReceivedEventHandler, logger)
	migrations := sqlite.NewMigrations(db)
	service := NewService(application, server, metricsServer, downloader, receivedEventSubscriber, migrations)
	integrationService := IntegrationService{
		Service: service,
	}
	return integrationService, func() {
		cleanup()
	}, nil
}

func BuildTestAdapters(contextContext context.Context, tb testing.TB) (sqlite.TestedItems, func(), error) {
	configConfig, err := newTestAdaptersConfig(tb)
	if err != nil {
		return sqlite.TestedItems{}, nil, err
	}
	logger, err := newLogger(configConfig)
	if err != nil {
		return sqlite.TestedItems{}, nil, err
	}
	db, cleanup, err := newSqliteDB(configConfig, logger)
	if err != nil {
		return sqlite.TestedItems{}, nil, err
	}
	diBuildTransactionSqliteAdaptersDependencies := buildTransactionSqliteAdaptersDependencies{}
	genericAdaptersFactoryFn := newTestAdaptersFactoryFn(diBuildTransactionSqliteAdaptersDependencies)
	genericTransactionProvider := sqlite.NewTestTransactionProvider(db, genericAdaptersFactoryFn)
	migrations := sqlite.NewMigrations(db)
	testedItems := sqlite.TestedItems{
		TransactionProvider: genericTransactionProvider,
		Migrations:          migrations,
	}
	return testedItems, func() {
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
	publicKeyRepository, err := sqlite.NewPublicKeyRepository(tx)
	if err != nil {
		return app.Adapters{}, err
	}
	processedEventRepository, err := sqlite.NewProcessedEventRepository(tx)
	if err != nil {
		return app.Adapters{}, err
	}
	appAdapters := app.Adapters{
		Accounts:        accountRepository,
		Sessions:        sessionRepository,
		PublicKeys:      publicKeyRepository,
		ProcessedEvents: processedEventRepository,
	}
	return appAdapters, nil
}

func buildTestTransactionSqliteAdapters(db *sql.DB, tx *sql.Tx, diBuildTransactionSqliteAdaptersDependencies buildTransactionSqliteAdaptersDependencies) (sqlite.TestAdapters, error) {
	sessionRepository, err := sqlite.NewSessionRepository(tx)
	if err != nil {
		return sqlite.TestAdapters{}, err
	}
	accountRepository, err := sqlite.NewAccountRepository(tx)
	if err != nil {
		return sqlite.TestAdapters{}, err
	}
	publicKeyRepository, err := sqlite.NewPublicKeyRepository(tx)
	if err != nil {
		return sqlite.TestAdapters{}, err
	}
	processedEventRepository, err := sqlite.NewProcessedEventRepository(tx)
	if err != nil {
		return sqlite.TestAdapters{}, err
	}
	testAdapters := sqlite.TestAdapters{
		SessionRepository:        sessionRepository,
		AccountRepository:        accountRepository,
		PublicKeyRepository:      publicKeyRepository,
		ProcessedEventRepository: processedEventRepository,
	}
	return testAdapters, nil
}

// wire.go:

type IntegrationService struct {
	Service Service
}

func newTestAdaptersConfig(tb testing.TB) (config.Config, error) {
	return config.NewConfig(fixtures.SomeString(), fixtures.SomeString(), config.EnvironmentDevelopment, logging.LevelDebug, fixtures.SomeString(), fixtures.SomeString(), fixtures.SomeFile(tb))
}

type buildTransactionSqliteAdaptersDependencies struct {
}

var downloaderSet = wire.NewSet(app.NewDownloader)
