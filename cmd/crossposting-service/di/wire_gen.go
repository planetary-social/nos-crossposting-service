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
	"github.com/planetary-social/nos-crossposting-service/migrations"
	"github.com/planetary-social/nos-crossposting-service/service/adapters"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/memorypubsub"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/mocks"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/prometheus"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/twitter"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/content"
	"github.com/planetary-social/nos-crossposting-service/service/ports/http"
	"github.com/planetary-social/nos-crossposting-service/service/ports/http/frontend"
	memorypubsub2 "github.com/planetary-social/nos-crossposting-service/service/ports/memorypubsub"
	"github.com/planetary-social/nos-crossposting-service/service/ports/sqlitepubsub"
	"github.com/planetary-social/nos-crossposting-service/service/ports/timer"
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
	diBuildTransactionSqliteAdaptersDependencies := buildTransactionSqliteAdaptersDependencies{
		Logger: logger,
	}
	genericAdaptersFactoryFn := newAdaptersFactoryFn(diBuildTransactionSqliteAdaptersDependencies)
	genericTransactionProvider := sqlite.NewTransactionProvider(db, genericAdaptersFactoryFn)
	prometheusPrometheus, err := prometheus.NewPrometheus(logger)
	if err != nil {
		cleanup()
		return Service{}, nil, err
	}
	getSessionAccountHandler := app.NewGetSessionAccountHandler(genericTransactionProvider, logger, prometheusPrometheus)
	getAccountPublicKeysHandler := app.NewGetAccountPublicKeysHandler(genericTransactionProvider, logger, prometheusPrometheus)
	twitterTwitter := twitter.NewTwitter(configConfig, logger, prometheusPrometheus)
	developmentTwitter := twitter.NewDevelopmentTwitter(logger)
	appTwitter := selectTwitterAdapterDependingOnConfig(configConfig, twitterTwitter, developmentTwitter)
	twitterAccountDetailsCache := adapters.NewTwitterAccountDetailsCache()
	getTwitterAccountDetailsHandler := app.NewGetTwitterAccountDetailsHandler(genericTransactionProvider, appTwitter, twitterAccountDetailsCache, logger, prometheusPrometheus)
	idGenerator := adapters.NewIDGenerator()
	loginOrRegisterHandler := app.NewLoginOrRegisterHandler(genericTransactionProvider, idGenerator, idGenerator, logger, prometheusPrometheus)
	logoutHandler := app.NewLogoutHandler(genericTransactionProvider, logger, prometheusPrometheus)
	linkPublicKeyHandler := app.NewLinkPublicKeyHandler(genericTransactionProvider, logger, prometheusPrometheus)
	unlinkPublicKeyHandler := app.NewUnlinkPublicKeyHandler(genericTransactionProvider, logger, prometheusPrometheus)
	pubSub := sqlite.NewPubSub(db, logger)
	subscriber := sqlite.NewSubscriber(pubSub, db)
	updateMetricsHandler := app.NewUpdateMetricsHandler(genericTransactionProvider, subscriber, logger, prometheusPrometheus)
	application := app.Application{
		GetSessionAccount:        getSessionAccountHandler,
		GetAccountPublicKeys:     getAccountPublicKeysHandler,
		GetTwitterAccountDetails: getTwitterAccountDetailsHandler,
		LoginOrRegister:          loginOrRegisterHandler,
		Logout:                   logoutHandler,
		LinkPublicKey:            linkPublicKeyHandler,
		UnlinkPublicKey:          unlinkPublicKeyHandler,
		UpdateMetrics:            updateMetricsHandler,
	}
	frontendFileSystem, err := frontend.NewFrontendFileSystem()
	if err != nil {
		cleanup()
		return Service{}, nil, err
	}
	server := http.NewServer(configConfig, application, logger, frontendFileSystem)
	metricsServer := http.NewMetricsServer(prometheusPrometheus, configConfig, logger)
	receivedEventPubSub := memorypubsub.NewReceivedEventPubSub()
	purplePages, err := adapters.NewPurplePages(contextContext, logger, prometheusPrometheus)
	if err != nil {
		cleanup()
		return Service{}, nil, err
	}
	relaySource := adapters.NewRelaySource(logger, purplePages)
	relayEventDownloader := adapters.NewRelayEventDownloader(contextContext, logger, prometheusPrometheus)
	downloader := app.NewDownloader(genericTransactionProvider, receivedEventPubSub, logger, prometheusPrometheus, relaySource, relayEventDownloader)
	transformer := content.NewTransformer()
	tweetGenerator := domain.NewTweetGenerator(transformer)
	processReceivedEventHandler := app.NewProcessReceivedEventHandler(genericTransactionProvider, tweetGenerator, logger, prometheusPrometheus)
	receivedEventSubscriber := memorypubsub2.NewReceivedEventSubscriber(receivedEventPubSub, processReceivedEventHandler, logger)
	currentTimeProvider := adapters.NewCurrentTimeProvider()
	sendTweetHandler := app.NewSendTweetHandler(genericTransactionProvider, appTwitter, currentTimeProvider, logger, prometheusPrometheus)
	tweetCreatedEventSubscriber := sqlitepubsub.NewTweetCreatedEventSubscriber(sendTweetHandler, subscriber, logger)
	metrics := timer.NewMetrics(application, logger)
	migrationsStorage, err := sqlite.NewMigrationsStorage(db)
	if err != nil {
		cleanup()
		return Service{}, nil, err
	}
	runner := migrations.NewRunner(migrationsStorage, logger)
	migrationFns := sqlite.NewMigrationFns(db, pubSub)
	migrationsMigrations, err := sqlite.NewMigrations(migrationFns)
	if err != nil {
		cleanup()
		return Service{}, nil, err
	}
	loggingMigrationsProgressCallback := adapters.NewLoggingMigrationsProgressCallback(logger)
	service := NewService(application, server, metricsServer, downloader, receivedEventSubscriber, tweetCreatedEventSubscriber, metrics, runner, migrationsMigrations, loggingMigrationsProgressCallback)
	return service, func() {
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
	diBuildTransactionSqliteAdaptersDependencies := buildTransactionSqliteAdaptersDependencies{
		Logger: logger,
	}
	genericAdaptersFactoryFn := newTestAdaptersFactoryFn(diBuildTransactionSqliteAdaptersDependencies)
	genericTransactionProvider := sqlite.NewTestTransactionProvider(db, genericAdaptersFactoryFn)
	pubSub := sqlite.NewPubSub(db, logger)
	subscriber := sqlite.NewSubscriber(pubSub, db)
	migrationsStorage, err := sqlite.NewMigrationsStorage(db)
	if err != nil {
		cleanup()
		return sqlite.TestedItems{}, nil, err
	}
	runner := migrations.NewRunner(migrationsStorage, logger)
	migrationFns := sqlite.NewMigrationFns(db, pubSub)
	migrationsMigrations, err := sqlite.NewMigrations(migrationFns)
	if err != nil {
		cleanup()
		return sqlite.TestedItems{}, nil, err
	}
	loggingMigrationsProgressCallback := adapters.NewLoggingMigrationsProgressCallback(logger)
	testedItems := sqlite.TestedItems{
		TransactionProvider:        genericTransactionProvider,
		Subscriber:                 subscriber,
		MigrationsStorage:          migrationsStorage,
		PubSub:                     pubSub,
		MigrationsRunner:           runner,
		Migrations:                 migrationsMigrations,
		MigrationsProgressCallback: loggingMigrationsProgressCallback,
	}
	return testedItems, func() {
		cleanup()
	}, nil
}

func BuildTestApplication(tb testing.TB) (TestApplication, error) {
	accountRepository, err := mocks.NewAccountRepository()
	if err != nil {
		return TestApplication{}, err
	}
	sessionRepository, err := mocks.NewSessionRepository()
	if err != nil {
		return TestApplication{}, err
	}
	publicKeyRepository, err := mocks.NewPublicKeyRepository()
	if err != nil {
		return TestApplication{}, err
	}
	processedEventRepository, err := mocks.NewProcessedEventRepository()
	if err != nil {
		return TestApplication{}, err
	}
	userTokensRepository, err := mocks.NewUserTokensRepository()
	if err != nil {
		return TestApplication{}, err
	}
	publisher := mocks.NewPublisher()
	appAdapters := app.Adapters{
		Accounts:        accountRepository,
		Sessions:        sessionRepository,
		PublicKeys:      publicKeyRepository,
		ProcessedEvents: processedEventRepository,
		UserTokens:      userTokensRepository,
		Publisher:       publisher,
	}
	transactionProvider := mocks.NewTransactionProvider(appAdapters)
	mocksTwitter := mocks.NewTwitter()
	currentTimeProvider := mocks.NewCurrentTimeProvider()
	logger := fixtures.TestLogger(tb)
	prometheusPrometheus, err := prometheus.NewPrometheus(logger)
	if err != nil {
		return TestApplication{}, err
	}
	sendTweetHandler := app.NewSendTweetHandler(transactionProvider, mocksTwitter, currentTimeProvider, logger, prometheusPrometheus)
	testApplication := TestApplication{
		SendTweetHandler:     sendTweetHandler,
		CurrentTimeProvider:  currentTimeProvider,
		UserTokensRepository: userTokensRepository,
		Twitter:              mocksTwitter,
	}
	return testApplication, nil
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
	userTokensRepository, err := sqlite.NewUserTokensRepository(tx)
	if err != nil {
		return app.Adapters{}, err
	}
	logger := diBuildTransactionSqliteAdaptersDependencies.Logger
	pubSub := sqlite.NewPubSub(db, logger)
	publisher := sqlite.NewPublisher(pubSub, tx)
	appAdapters := app.Adapters{
		Accounts:        accountRepository,
		Sessions:        sessionRepository,
		PublicKeys:      publicKeyRepository,
		ProcessedEvents: processedEventRepository,
		UserTokens:      userTokensRepository,
		Publisher:       publisher,
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
	userTokensRepository, err := sqlite.NewUserTokensRepository(tx)
	if err != nil {
		return sqlite.TestAdapters{}, err
	}
	logger := diBuildTransactionSqliteAdaptersDependencies.Logger
	pubSub := sqlite.NewPubSub(db, logger)
	publisher := sqlite.NewPublisher(pubSub, tx)
	testAdapters := sqlite.TestAdapters{
		SessionRepository:        sessionRepository,
		AccountRepository:        accountRepository,
		PublicKeyRepository:      publicKeyRepository,
		ProcessedEventRepository: processedEventRepository,
		UserTokensRepository:     userTokensRepository,
		Publisher:                publisher,
	}
	return testAdapters, nil
}

// wire.go:

type TestApplication struct {
	SendTweetHandler *app.SendTweetHandler

	CurrentTimeProvider  *mocks.CurrentTimeProvider
	UserTokensRepository *mocks.UserTokensRepository
	Twitter              *mocks.Twitter
}

func newTestAdaptersConfig(tb testing.TB) (config.Config, error) {
	return config.NewConfig(fixtures.SomeString(), fixtures.SomeString(), config.EnvironmentDevelopment, logging.LevelDebug, fixtures.SomeString(), fixtures.SomeString(), fixtures.SomeFile(tb), fixtures.SomeString())
}

type buildTransactionSqliteAdaptersDependencies struct {
	Logger logging.Logger
}

var downloaderSet = wire.NewSet(app.NewDownloader)

var tweetGeneratorSet = wire.NewSet(content.NewTransformer, domain.NewTweetGenerator, wire.Bind(new(app.TweetGenerator), new(*domain.TweetGenerator)))
