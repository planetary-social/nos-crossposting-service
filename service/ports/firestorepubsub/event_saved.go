package firestorepubsub

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/firestore"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

const (
	gatherMetricsEvery = 2 * time.Minute
)

type ProcessSavedEventHandler interface {
	Handle(ctx context.Context, cmd app.ProcessSavedEvent) error
}

type Metrics interface {
	ReportSubscriptionQueueLength(topic string, n int)
}

type FirestoreSubscriber interface {
	message.Subscriber
	QueueLength(topic string) (int, error)
}

type EventSavedSubscriber struct {
	subscriber FirestoreSubscriber
	handler    ProcessSavedEventHandler
	metrics    Metrics
	logger     logging.Logger
}

func NewEventSavedSubscriber(
	subscriber FirestoreSubscriber,
	handler ProcessSavedEventHandler,
	metrics Metrics,
	logger logging.Logger,
) *EventSavedSubscriber {
	return &EventSavedSubscriber{
		subscriber: subscriber,
		handler:    handler,
		metrics:    metrics,
		logger:     logger.New("eventSavedSubscriber"),
	}
}

func (p *EventSavedSubscriber) Run(ctx context.Context) error {
	go p.gatherMetricsLoop(ctx)

	ch, err := p.subscriber.Subscribe(ctx, firestore.PubsubTopicEventSaved)
	if err != nil {
		return errors.Wrap(err, "error subscribing")
	}

	for msg := range ch {
		go p.handleMessage(ctx, msg)
	}
	return nil
}

func (p *EventSavedSubscriber) gatherMetricsLoop(ctx context.Context) {
	for {
		if err := p.gatherMetrics(); err != nil {
			p.logger.Error().WithError(err).Message("error gathering metrics")
		}

		select {
		case <-time.After(gatherMetricsEvery):
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (p *EventSavedSubscriber) gatherMetrics() error {
	n, err := p.subscriber.QueueLength(firestore.PubsubTopicEventSaved)
	if err != nil {
		return errors.Wrap(err, "error checking queue length")
	}

	p.metrics.ReportSubscriptionQueueLength(firestore.PubsubTopicEventSaved, n)
	return nil
}

func (p *EventSavedSubscriber) handleMessage(ctx context.Context, msg *message.Message) {
	if err := p.runHandler(ctx, msg); err != nil {
		p.logger.Error().WithError(err).Message("error handling a message")
		msg.Nack()
	} else {
		msg.Ack()
	}
}

func (p *EventSavedSubscriber) runHandler(ctx context.Context, msg *message.Message) error {
	var payload firestore.EventSavedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return errors.Wrap(err, "error unmarshaling event payload")
	}

	eventId, err := domain.NewEventId(payload.EventId)
	if err != nil {
		return errors.Wrap(err, "error creating event id")
	}

	cmd := app.NewProcessSavedEvent(eventId)
	if err := p.handler.Handle(ctx, cmd); err != nil {
		return errors.Wrap(err, "error calling the handler")
	}

	return nil
}
