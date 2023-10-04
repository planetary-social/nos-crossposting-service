package app

import (
	"context"
	"fmt"

	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

type SaveReceivedEvent struct {
	relay domain.RelayAddress
	event domain.Event
}

func NewSaveReceivedEvent(relay domain.RelayAddress, event domain.Event) SaveReceivedEvent {
	return SaveReceivedEvent{relay: relay, event: event}
}

type SaveReceivedEventHandler struct {
	eventWasAlreadySavedCache EventWasAlreadySavedCache
	transactionProvider       TransactionProvider
	logger                    logging.Logger
	metrics                   Metrics
}

func NewSaveReceivedEventHandler(
	eventWasAlreadySavedCache EventWasAlreadySavedCache,
	transactionProvider TransactionProvider,
	logger logging.Logger,
	metrics Metrics,
) *SaveReceivedEventHandler {
	return &SaveReceivedEventHandler{
		eventWasAlreadySavedCache: eventWasAlreadySavedCache,
		transactionProvider:       transactionProvider,
		logger:                    logger.New("saveReceivedEventHandler"),
		metrics:                   metrics,
	}
}

func (h *SaveReceivedEventHandler) Handle(ctx context.Context, cmd SaveReceivedEvent) (err error) {
	defer h.metrics.StartApplicationCall("saveReceivedEvent").End(&err)

	if !domain.ShouldDownloadEventKind(cmd.event.Kind()) {
		return fmt.Errorf("event '%s' shouldn't have been downloaded", cmd.event.String())
	}

	if h.eventWasAlreadySavedCache.EventWasAlreadySaved(cmd.event.Id()) {
		return nil
	}

	h.logger.Debug().
		WithField("relay", cmd.relay.String()).
		WithField("event.id", cmd.event.Id().Hex()).
		WithField("event.kind", cmd.event.Kind().Int()).
		WithField("size", len(cmd.event.Raw())).
		WithField("number_of_tags", len(cmd.event.Tags())).
		Message("saving received event")

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
