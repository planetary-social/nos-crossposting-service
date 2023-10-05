package app

import (
	"context"
	"fmt"

	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

type ProcessReceivedEvent struct {
	relay domain.RelayAddress
	event domain.Event
}

func NewProcessReceivedEvent(relay domain.RelayAddress, event domain.Event) ProcessReceivedEvent {
	return ProcessReceivedEvent{relay: relay, event: event}
}

type SaveReceivedEventHandler struct {
	transactionProvider TransactionProvider
	logger              logging.Logger
	metrics             Metrics
}

func NewSaveReceivedEventHandler(
	transactionProvider TransactionProvider,
	logger logging.Logger,
	metrics Metrics,
) *SaveReceivedEventHandler {
	return &SaveReceivedEventHandler{
		transactionProvider: transactionProvider,
		logger:              logger.New("processReceivedEventHandler"),
		metrics:             metrics,
	}
}

func (h *SaveReceivedEventHandler) Handle(ctx context.Context, cmd ProcessReceivedEvent) (err error) {
	defer h.metrics.StartApplicationCall("processReceivedEvent").End(&err)

	if !domain.ShouldDownloadEventKind(cmd.event.Kind()) {
		return fmt.Errorf("event '%s' shouldn't have been downloaded", cmd.event.String())
	}

	h.logger.Trace().
		WithField("relay", cmd.relay.String()).
		WithField("event.id", cmd.event.Id().Hex()).
		WithField("event.kind", cmd.event.Kind().Int()).
		WithField("size", len(cmd.event.Raw())).
		WithField("number_of_tags", len(cmd.event.Tags())).
		Message("processing received event")

	// todo

	//if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
	//	exists, err := adapters.Events.Exists(ctx, cmd.event.Id())
	//	if err != nil {
	//		return errors.Wrap(err, "error checking if event exists")
	//	}
	//
	//	if exists {
	//		h.eventWasAlreadySavedCache.MarkEventAsAlreadySaved(cmd.event.Id())
	//		return nil
	//	}
	//
	//	if err := adapters.Events.Save(cmd.event); err != nil {
	//		return errors.Wrap(err, "error saving the event")
	//	}
	//
	//	//if err := adapters.Publisher.PublishEventSaved(ctx, cmd.event.Id()); err != nil {
	//	//	return errors.Wrap(err, "error publishing")
	//	//}
	//
	//	return nil
	//}); err != nil {
	//	return errors.Wrap(err, "transaction error")
	//}

	return nil
}
