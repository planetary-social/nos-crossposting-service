package di

import (
	"context"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/nos-crossposting-service/service/adapters"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/ports/firestorepubsub"
	"github.com/planetary-social/nos-crossposting-service/service/ports/http"
	"github.com/planetary-social/nos-crossposting-service/service/ports/memorypubsub"
)

type Service struct {
	app                       app.Application
	server                    http.Server
	metricsServer             http.MetricsServer
	downloader                *app.Downloader
	receivedEventSubscriber   *memorypubsub.ReceivedEventSubscriber
	eventSavedSubscriber      *firestorepubsub.EventSavedSubscriber
	eventWasAlreadySavedCache *adapters.MemoryEventWasAlreadySavedCache
}

func NewService(
	app app.Application,
	server http.Server,
	metricsServer http.MetricsServer,
	downloader *app.Downloader,
	receivedEventSubscriber *memorypubsub.ReceivedEventSubscriber,
	eventSavedSubscriber *firestorepubsub.EventSavedSubscriber,
	eventWasAlreadySavedCache *adapters.MemoryEventWasAlreadySavedCache,
) Service {
	return Service{
		app:                       app,
		server:                    server,
		metricsServer:             metricsServer,
		downloader:                downloader,
		receivedEventSubscriber:   receivedEventSubscriber,
		eventSavedSubscriber:      eventSavedSubscriber,
		eventWasAlreadySavedCache: eventWasAlreadySavedCache,
	}
}

func (s Service) App() app.Application {
	return s.app
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
		errCh <- errors.Wrap(s.eventSavedSubscriber.Run(ctx), "event saved subscriber error")
	}()

	runners++
	go func() {
		errCh <- errors.Wrap(s.eventWasAlreadySavedCache.Run(ctx), "event was already saved cache error")
	}()

	var err error
	for i := 0; i < runners; i++ {
		err = multierror.Append(err, errors.Wrap(<-errCh, "error returned by runner"))
		cancel()
	}

	return err
}
