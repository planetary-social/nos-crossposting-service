// Package pubsub receives internal events.
package memorypubsub

import (
	"context"

	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/memorypubsub"
	"github.com/planetary-social/nos-crossposting-service/service/app"
)

type SaveReceivedEventHandler interface {
	Handle(ctx context.Context, cmd app.ProcessReceivedEvent) error
}

type ReceivedEventSubscriber struct {
	pubsub  *memorypubsub.ReceivedEventPubSub
	handler SaveReceivedEventHandler
	logger  logging.Logger
}

func NewReceivedEventSubscriber(
	pubsub *memorypubsub.ReceivedEventPubSub,
	handler SaveReceivedEventHandler,
	logger logging.Logger,
) *ReceivedEventSubscriber {
	return &ReceivedEventSubscriber{
		pubsub:  pubsub,
		handler: handler,
		logger:  logger.New("receivedEventSubscriber"),
	}
}

func (p *ReceivedEventSubscriber) Run(ctx context.Context) error {
	for v := range p.pubsub.Subscribe(ctx) {
		cmd := app.NewProcessReceivedEvent(v.Relay(), v.Event())
		if err := p.handler.Handle(ctx, cmd); err != nil {
			p.logger.Error().
				WithError(err).
				WithField("relay", v.Relay()).
				WithField("event", v.Event()).
				Message("error handling a received event")
		}
	}
	return nil
}
