package app

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

type EventOrEOSEOrError struct {
	event domain.Event
	eose  bool
	err   error
}

func NewEventOrEOSEOrErrorWithEvent(event domain.Event) EventOrEOSEOrError {
	return EventOrEOSEOrError{event: event}
}

func NewEventOrEOSEOrErrorWithError(err error) EventOrEOSEOrError {
	return EventOrEOSEOrError{err: err}
}

func NewEventOrEOSEOrErrorWithEOSE() EventOrEOSEOrError {
	return EventOrEOSEOrError{eose: true}
}

func (e *EventOrEOSEOrError) Event() domain.Event {
	return e.event
}

func (e *EventOrEOSEOrError) EOSE() bool {
	return e.eose
}

func (e *EventOrEOSEOrError) Err() error {
	return e.err
}

type GetEventsHandler struct {
	transactionProvider     TransactionProvider
	receivedEventSubscriber ReceivedEventSubscriber
	metrics                 Metrics
}

func NewGetEventsHandler(
	transactionProvider TransactionProvider,
	receivedEventSubscriber ReceivedEventSubscriber,
	metrics Metrics,
) *GetEventsHandler {
	return &GetEventsHandler{
		transactionProvider:     transactionProvider,
		receivedEventSubscriber: receivedEventSubscriber,
		metrics:                 metrics,
	}
}

func (h *GetEventsHandler) Handle(ctx context.Context, filters domain.Filters) <-chan EventOrEOSEOrError {
	defer h.metrics.StartApplicationCall("getEvents").End(nil)

	ch := make(chan EventOrEOSEOrError)
	go h.send(ctx, filters, ch)
	return ch
}

func (h *GetEventsHandler) send(ctx context.Context, filters domain.Filters, ch chan<- EventOrEOSEOrError) {
	defer close(ch)

	receivedEvents := h.receivedEventSubscriber.Subscribe(ctx)

	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		for eventOrErr := range adapters.Events.GetEvents(ctx, filters) {
			if err := eventOrErr.Err(); err != nil {
				return errors.Wrap(err, "repository returned an error")
			}

			select {
			case ch <- NewEventOrEOSEOrErrorWithEvent(eventOrErr.Event()):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}); err != nil {
		select {
		case ch <- NewEventOrEOSEOrErrorWithError(errors.Wrap(err, "transaction failed")):
		case <-ctx.Done():
		}
		return
	}

	select {
	case ch <- NewEventOrEOSEOrErrorWithEOSE():
	case <-ctx.Done():
		return
	}

	for receivedEvent := range receivedEvents {
		if filters.Match(receivedEvent.Event()) {
			select {
			case ch <- NewEventOrEOSEOrErrorWithEvent(receivedEvent.Event()):
			case <-ctx.Done():
				return
			}
		}
	}
}
