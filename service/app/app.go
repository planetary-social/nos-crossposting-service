package app

import (
	"context"
	"time"

	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/notifications"
)

type TransactionProvider interface {
	Transact(context.Context, func(context.Context, Adapters) error) error
}

type Adapters struct {
	Registrations RegistrationRepository
	Relays        RelayRepository
	PublicKeys    PublicKeyRepository
	Events        EventRepository
	Tags          TagRepository

	Publisher Publisher
}

type RegistrationRepository interface {
	Save(registration domain.Registration) error
}

type RelayRepository interface {
	GetRelays(ctx context.Context, updatedAfter time.Time) ([]domain.RelayAddress, error)
	GetPublicKeys(ctx context.Context, address domain.RelayAddress, updatedAfter time.Time) ([]domain.PublicKey, error)
}

type PublicKeyRepository interface {
	GetAPNSTokens(ctx context.Context, publicKey domain.PublicKey, savedAfter time.Time) ([]domain.APNSToken, error)
}

type EventRepository interface {
	Save(event domain.Event) error
	Get(ctx context.Context, id domain.EventId) (domain.Event, error)
	Exists(ctx context.Context, id domain.EventId) (bool, error)
	GetEvents(ctx context.Context, filters domain.Filters) <-chan EventOrError
	SaveNotificationForEvent(notification notifications.Notification) error
	GetNotifications(ctx context.Context, id domain.EventId) ([]notifications.Notification, error)
}

type TagRepository interface {
	Save(event domain.Event, tags []domain.EventTag) error
}

type Publisher interface {
	PublishEventSaved(ctx context.Context, id domain.EventId) error
}

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	SaveReceivedEvent *SaveReceivedEventHandler
	SaveRegistration  *SaveRegistrationHandler
}

type Queries struct {
	GetRelays        *GetRelaysHandler
	GetPublicKeys    *GetPublicKeysHandler
	GetTokens        *GetTokensHandler
	GetEvents        *GetEventsHandler
	GetNotifications *GetNotificationsHandler
}

type APNS interface {
	SendNotification(notification notifications.Notification) error
}

type EventOrError struct {
	event domain.Event
	err   error
}

func NewEventOrErrorWithEvent(event domain.Event) EventOrError {
	return EventOrError{event: event}
}

func NewEventOrErrorWithError(err error) EventOrError {
	return EventOrError{err: err}
}

func (e *EventOrError) Event() domain.Event {
	return e.event
}

func (e *EventOrError) Err() error {
	return e.err
}

type ReceivedEvent struct {
	relay domain.RelayAddress
	event domain.Event
}

func NewReceivedEvent(relay domain.RelayAddress, event domain.Event) ReceivedEvent {
	return ReceivedEvent{relay: relay, event: event}
}

func (r ReceivedEvent) Relay() domain.RelayAddress {
	return r.relay
}

func (r ReceivedEvent) Event() domain.Event {
	return r.event
}

type ReceivedEventSubscriber interface {
	Subscribe(ctx context.Context) <-chan ReceivedEvent
}

type Metrics interface {
	StartApplicationCall(handlerName string) ApplicationCall
	MeasureRelayDownloadersState(n int, state RelayDownloaderState)
}

type ApplicationCall interface {
	// End accepts a pointer so that you can defer this call without wrapping it
	// in an anonymous function
	End(err *error)
}

type EventWasAlreadySavedCache interface {
	MarkEventAsAlreadySaved(id domain.EventId)
	EventWasAlreadySaved(id domain.EventId) bool
}
