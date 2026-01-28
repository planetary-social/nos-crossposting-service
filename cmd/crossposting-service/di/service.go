package di

import (
	"context"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/nos-crossposting-service/internal/goroutine"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/migrations"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/ports/http"
	"github.com/planetary-social/nos-crossposting-service/service/ports/memorypubsub"
	"github.com/planetary-social/nos-crossposting-service/service/ports/sqlitepubsub"
	"github.com/planetary-social/nos-crossposting-service/service/ports/timer"
)

type Service struct {
	app                         app.Application
	server                      http.Server
	metricsServer               http.MetricsServer
	downloader                  *app.Downloader
	receivedEventSubscriber     *memorypubsub.ReceivedEventSubscriber
	tweetCreatedEventSubscriber *sqlitepubsub.TweetCreatedEventSubscriber
	metricsTimer                *timer.Metrics
	migrationsRunner            *migrations.Runner
	migrations                  migrations.Migrations
	migrationsProgressCallback  migrations.ProgressCallback
	vanishSubscriber            *app.VanishSubscriber
	logger                      logging.Logger
}

func NewService(
	app app.Application,
	server http.Server,
	metricsServer http.MetricsServer,
	downloader *app.Downloader,
	receivedEventSubscriber *memorypubsub.ReceivedEventSubscriber,
	tweetCreatedEventSubscriber *sqlitepubsub.TweetCreatedEventSubscriber,
	metricsTimer *timer.Metrics,
	migrationsRunner *migrations.Runner,
	migrations migrations.Migrations,
	migrationsProgressCallback migrations.ProgressCallback,
	vanishSubscriber *app.VanishSubscriber,
	logger logging.Logger,
) Service {
	return Service{
		app:                         app,
		server:                      server,
		metricsServer:               metricsServer,
		downloader:                  downloader,
		receivedEventSubscriber:     receivedEventSubscriber,
		tweetCreatedEventSubscriber: tweetCreatedEventSubscriber,
		metricsTimer:                metricsTimer,
		migrationsRunner:            migrationsRunner,
		migrations:                  migrations,
		migrationsProgressCallback:  migrationsProgressCallback,
		vanishSubscriber:            vanishSubscriber,
		logger:                      logger.New("service"),
	}
}

func (s Service) App() app.Application {
	return s.app
}

func (s Service) ExecuteMigrations(ctx context.Context) error {
	return s.migrationsRunner.Run(ctx, s.migrations, s.migrationsProgressCallback)
}

func (s Service) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error)
	runners := 0

	// All goroutines are wrapped with panic recovery. If any goroutine panics,
	// the panic is logged with a stack trace and converted to an error, which
	// triggers context cancellation and graceful shutdown of all other goroutines.
	// This prevents zombie goroutines from running after a panic.

	runners++
	goroutine.Run(errCh, s.logger, "http-server", func() error {
		return s.server.ListenAndServe(ctx)
	})

	runners++
	goroutine.Run(errCh, s.logger, "metrics-server", func() error {
		return s.metricsServer.ListenAndServe(ctx)
	})

	runners++
	goroutine.Run(errCh, s.logger, "downloader", func() error {
		return s.downloader.Run(ctx)
	})

	runners++
	goroutine.Run(errCh, s.logger, "received-event-subscriber", func() error {
		return s.receivedEventSubscriber.Run(ctx)
	})

	runners++
	goroutine.Run(errCh, s.logger, "tweet-created-event-subscriber", func() error {
		return s.tweetCreatedEventSubscriber.Run(ctx)
	})

	runners++
	goroutine.Run(errCh, s.logger, "metrics-timer", func() error {
		return s.metricsTimer.Run(ctx)
	})

	runners++
	goroutine.Run(errCh, s.logger, "vanish-subscriber", func() error {
		return s.vanishSubscriber.Run(ctx)
	})

	var err error
	for i := 0; i < runners; i++ {
		err = multierror.Append(err, errors.Wrap(<-errCh, "error returned by runner"))
		cancel()
	}

	return err
}
