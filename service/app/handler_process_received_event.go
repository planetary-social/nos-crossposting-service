package app

import (
	"context"
	"fmt"

	"github.com/boreq/errors"
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

type ProcessReceivedEventHandler struct {
	transactionProvider TransactionProvider
	tweetGenerator      TweetGenerator
	logger              logging.Logger
	metrics             Metrics
}

func NewProcessReceivedEventHandler(
	transactionProvider TransactionProvider,
	tweetGenerator TweetGenerator,
	logger logging.Logger,
	metrics Metrics,
) *ProcessReceivedEventHandler {
	return &ProcessReceivedEventHandler{
		transactionProvider: transactionProvider,
		tweetGenerator:      tweetGenerator,
		logger:              logger.New("processReceivedEventHandler"),
		metrics:             metrics,
	}
}

func (h *ProcessReceivedEventHandler) Handle(ctx context.Context, cmd ProcessReceivedEvent) (err error) {
	defer h.metrics.StartApplicationCall("processReceivedEvent").End(&err)

	if !domain.ShouldDownloadEventKind(cmd.event.Kind()) {
		return fmt.Errorf("event '%s' shouldn't have been downloaded", cmd.event.String())
	}

	event := cmd.event

	h.logger.Trace().
		WithField("relay", cmd.relay.String()).
		WithField("event.id", event.Id().Hex()).
		WithField("event.kind", event.Kind().Int()).
		WithField("size", len(event.Raw())).
		WithField("number_of_tags", len(event.Tags())).
		Message("processing received event")

	tweets, err := h.tweetGenerator.Generate(event)
	if err != nil {
		return errors.Wrapf(err, "error generating tweets for event '%s'", event.Id())
	}

	if thereIsNothingToDo := len(tweets) == 0; thereIsNothingToDo {
		return nil
	}

	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		linkedPublicKeys, err := adapters.PublicKeys.ListByPublicKey(event.PublicKey())
		if err != nil {
			return errors.Wrap(err, "error checking if event exists")
		}

		for _, linkedPublicKey := range linkedPublicKeys {
			if h.eventWasCreatedBeforePublicKeyWasLinked(event, linkedPublicKey) {
				continue
			}

			account, err := adapters.Accounts.GetByAccountID(linkedPublicKey.AccountID())
			if err != nil {
				return errors.Wrapf(err, "error getting an account '%s'", linkedPublicKey.AccountID().String())
			}

			wasProcessed, err := adapters.ProcessedEvents.WasProcessed(event.Id(), account.TwitterID())
			if err != nil {
				return errors.Wrap(err, "error checking if event was processed")
			}

			if wasProcessed {
				continue
			}

			if err := adapters.ProcessedEvents.Save(event.Id(), account.TwitterID()); err != nil {
				return errors.Wrap(err, "error saving that event was processed")
			}

			for _, tweet := range tweets {
				if err := adapters.Publisher.PublishTweetCreated(account.AccountID(), tweet); err != nil {
					return errors.Wrap(err, "error publishing tweet created event")
				}
			}
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction error")
	}

	return nil
}

func (h *ProcessReceivedEventHandler) eventWasCreatedBeforePublicKeyWasLinked(event domain.Event, linkedPublicKey *domain.LinkedPublicKey) bool {
	return event.CreatedAt().Before(linkedPublicKey.CreatedAt())
}
