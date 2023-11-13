package di

import (
	"context"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
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

	runners++
	go func() {
		errCh <- errors.Wrap(s.server.ListenAndServe(ctx), "server error")
	}()

	runners++
	go func() {
		errCh <- errors.Wrap(s.metricsServer.ListenAndServe(ctx), "metrics server error")
	}()

	runners++
	go func() {
		errCh <- errors.Wrap(s.downloader.Run(ctx), "downloader error")
	}()

	runners++
	go func() {
		errCh <- errors.Wrap(s.receivedEventSubscriber.Run(ctx), "received event subscriber error")
	}()

	runners++
	go func() {
		errCh <- errors.Wrap(s.tweetCreatedEventSubscriber.Run(ctx), "tweet created event subscriber error")
	}()

	runners++
	go func() {
		errCh <- errors.Wrap(s.metricsTimer.Run(ctx), "metrics timer error")
	}()

	var err error
	for i := 0; i < runners; i++ {
		err = multierror.Append(err, errors.Wrap(<-errCh, "error returned by runner"))
		cancel()
	}

	return err
}
