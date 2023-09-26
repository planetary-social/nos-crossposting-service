package app

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/notifications"
)

const (
	tagBatchSize                       = 150
	apnsTokenBatchSize                 = 500
	onlySaveEventForEventsWithMoreTags = 500

	sendNotificationsToTokensYoungerThan = 6 * 30 * 24 * time.Hour
)

type ProcessSavedEvent struct {
	eventId domain.EventId
}

func NewProcessSavedEvent(eventId domain.EventId) ProcessSavedEvent {
	return ProcessSavedEvent{eventId: eventId}
}

type ProcessSavedEventHandler struct {
	transactionProvider TransactionProvider
	generator           *notifications.Generator
	apns                APNS
	logger              logging.Logger
	metrics             Metrics
}

func NewProcessSavedEventHandler(
	transactionProvider TransactionProvider,
	generator *notifications.Generator,
	apns APNS,
	logger logging.Logger,
	metrics Metrics,
) *ProcessSavedEventHandler {
	return &ProcessSavedEventHandler{
		transactionProvider: transactionProvider,
		generator:           generator,
		apns:                apns,
		logger:              logger.New("processSavedEventHandler"),
		metrics:             metrics,
	}
}

func (h *ProcessSavedEventHandler) Handle(ctx context.Context, cmd ProcessSavedEvent) (err error) {
	defer h.metrics.StartApplicationCall("processSavedEvent").End(&err)

	logger := h.logger.WithField("event.id", cmd.eventId.Hex())

	logger.Debug().Message("processing saved event")

	event, err := h.loadEvent(ctx, cmd)
	if err != nil {
		return errors.Wrap(err, "error loading event")
	}

	if len(event.Tags()) <= onlySaveEventForEventsWithMoreTags {
		if err := h.saveTags(ctx, event, logger); err != nil {
			return errors.Wrap(err, "error saving tags")
		}

		if err := h.generateSendAndSaveNotifications(ctx, event, logger); err != nil {
			return errors.Wrap(err, "error saving notifications")
		}
	}

	return nil
}

func (h *ProcessSavedEventHandler) loadEvent(ctx context.Context, cmd ProcessSavedEvent) (domain.Event, error) {
	var event domain.Event
	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		tmp, err := adapters.Events.Get(ctx, cmd.eventId)
		if err != nil {
			return errors.Wrap(err, "error getting the event from the database")
		}

		event = tmp
		return nil
	}); err != nil {
		return domain.Event{}, errors.Wrap(err, "event loading transaction error")
	}

	return event, nil
}

func (h *ProcessSavedEventHandler) saveTags(ctx context.Context, event domain.Event, logger logging.Logger) error {
	if len(event.Tags()) == 0 {
		return nil
	}

	tags := h.filterOutEmptyTags(event.Tags())

	logger.Debug().
		WithField("numberOfEventTags", len(event.Tags())).
		WithField("numberOfSavedTags", len(tags)).
		Message("saving tags")

	for _, batch := range internal.BatchesFromSlice(tags, tagBatchSize) {
		if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
			if err := adapters.Tags.Save(event, batch); err != nil {
				return errors.Wrap(err, "error saving the batch")
			}
			return nil
		}); err != nil {
			return errors.Wrap(err, "transaction error")
		}
	}

	return nil
}

func (h *ProcessSavedEventHandler) generateSendAndSaveNotifications(ctx context.Context, event domain.Event, logger logging.Logger) error {
	// todo this shouldn't send multiple notifications if the event is retried

	mentions, err := domain.GetMentionsFromTags(event.Tags())
	if err != nil {
		return errors.Wrap(err, "error getting mentions for this event")
	}

	var mentionToTokens map[domain.PublicKey][]domain.APNSToken
	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		mentionToTokens = make(map[domain.PublicKey][]domain.APNSToken) // transactions can run multiple times

		for _, mention := range mentions {
			tmp, err := adapters.PublicKeys.GetAPNSTokens(ctx, mention, time.Now().Add(-sendNotificationsToTokensYoungerThan))
			if err != nil {
				return errors.Wrap(err, "error getting the token")
			}
			if len(tmp) > 0 {
				mentionToTokens[mention] = append(mentionToTokens[mention], tmp...)
			}
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "token transaction error")
	}

	for mention, tokens := range mentionToTokens {
		logger.Debug().
			WithField("mention", mention.Hex()).
			WithField("numberOfTokens", len(tokens)).
			Message("sending notifications")

		for _, batch := range internal.BatchesFromSlice(tokens, apnsTokenBatchSize) {
			if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
				for _, token := range batch {
					notifications, err := h.generator.Generate(mention, token, event)
					if err != nil {
						return errors.Wrap(err, "error generating notifications")
					}

					for _, notification := range notifications {
						if err := h.apns.SendNotification(notification); err != nil {
							return errors.Wrap(err, "error sending a notification")
						}

						if err := adapters.Events.SaveNotificationForEvent(notification); err != nil {
							return errors.Wrap(err, "error saving notification")
						}
					}
				}

				return nil
			}); err != nil {
				return errors.Wrap(err, "transaction error")
			}
		}
	}

	return nil
}

// Since Firestore actually converts all paths to `slash/separated/strings` it
// doesn't understand the situation where things `accidently/end/with/a/slash/`
// as the last element is an empty string. Therefore, we can't save tags that
// have an empty value associated with them. This doesn't matter right now as we
// only search by `p` tags which always have a value associated with them.
//
// I am putting this in the application layer to make it more explicit.
func (h *ProcessSavedEventHandler) filterOutEmptyTags(tags []domain.EventTag) []domain.EventTag {
	var result []domain.EventTag
	for _, tag := range tags {
		if !tag.FirstValueIsAnEmptyString() {
			result = append(result, tag)
		}
	}
	return result
}
