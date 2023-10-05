package firestore

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	collectionEvents              = "events"
	collectionEventsNotifications = "notifications"

	eventFieldId        = "id"
	eventFieldPublicKey = "publicKey"
	eventFieldCreatedAt = "createdAt"
	eventFieldKind      = "kind"
	eventFieldRaw       = "raw"

	eventNotificationUUID    = "uuid"
	eventNotificationToken   = "token"
	eventNotificationPayload = "payload"
)

type EventRepository struct {
	client          *firestore.Client
	tx              *firestore.Transaction
	relayRepository *RelayRepository
	tagRepository   *TagRepository
}

func NewEventRepository(
	client *firestore.Client,
	tx *firestore.Transaction,
	relayRepository *RelayRepository,
	tagRepository *TagRepository,
) *EventRepository {
	return &EventRepository{
		client:          client,
		tx:              tx,
		relayRepository: relayRepository,
		tagRepository:   tagRepository,
	}
}

func (e *EventRepository) Save(event domain.Event) error {
	if err := e.saveUnderEvents(event); err != nil {
		return errors.Wrap(err, "error saving under events")
	}

	return nil
}

func (e *EventRepository) Exists(ctx context.Context, id domain.EventId) (bool, error) {
	_, err := e.client.Collection(collectionEvents).Doc(id.Hex()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, errors.Wrap(err, "error checking if document exists")
	}
	return true, nil
}

func (e *EventRepository) Get(ctx context.Context, id domain.EventId) (domain.Event, error) {
	doc, err := e.client.Collection(collectionEvents).Doc(id.Hex()).Get(ctx)
	if err != nil {
		return domain.Event{}, errors.Wrap(err, "error getting a doc")
	}

	event, err := e.readEvent(doc)
	if err != nil {
		return domain.Event{}, errors.Wrap(err, "error reading a doc")
	}

	return event, nil
}

func (e *EventRepository) saveUnderEvents(event domain.Event) error {
	eventDocPath := e.client.Collection(collectionEvents).Doc(event.Id().Hex())
	eventDocData := map[string]any{
		eventFieldId:        ensureType[string](event.Id().Hex()),
		eventFieldPublicKey: ensureType[string](event.PubKey().Hex()),
		eventFieldCreatedAt: ensureType[time.Time](event.CreatedAt()),
		eventFieldKind:      ensureType[int](event.Kind().Int()),
		eventFieldRaw:       ensureType[[]byte](event.Raw()),
	}
	if err := e.tx.Set(eventDocPath, eventDocData, firestore.MergeAll); err != nil {
		return errors.Wrap(err, "error updating the event doc")
	}

	return nil
}

func (e *EventRepository) GetEvents(ctx context.Context, filters domain.Filters) <-chan app.EventOrError {
	ch := make(chan app.EventOrError)
	go e.getEvents(ctx, filters, ch)
	return ch
}

func (e *EventRepository) getEvents(ctx context.Context, filters domain.Filters, ch chan<- app.EventOrError) {
	defer close(ch)

	events, err := e.loadEventsForFilters(ctx, filters)
	if err != nil {
		sendErr(ctx, ch, err)
		return
	}

	for _, event := range events {
		select {
		case ch <- app.NewEventOrErrorWithEvent(event):
		case <-ctx.Done():
		}
	}
}

func (e *EventRepository) loadEventsForFilters(ctx context.Context, filters domain.Filters) ([]domain.Event, error) {
	events := make(map[string]domain.Event)

	query := e.client.Collection(collectionEvents).Query

	// either the compound OR queries don't work with the simulator or they don't work at all
	// given how buggy the simulator has proven to be in the past maybe they work with the real firestore instance
	for _, filter := range filters.Filters() {
		if len(filter.Ids()) == 0 && len(filter.Kinds()) == 0 && len(filter.Authors()) == 0 && len(filter.Tags()) == 0 {
			if err := e.loadEvents(ctx, query, events, filter); err != nil {
				return nil, errors.Wrap(err, "error loading events")
			}
		} else {
			for _, v := range filter.Ids() {
				if err := e.loadEvents(ctx, query.Where(eventFieldId, "==", v.Hex()), events, filter); err != nil {
					return nil, errors.Wrapf(err, "error loading events for id filter '%s'", v.Hex())
				}
			}

			for _, v := range filter.Kinds() {
				if err := e.loadEvents(ctx, query.Where(eventFieldKind, "==", v.Int()), events, filter); err != nil {
					return nil, errors.Wrapf(err, "error loading events for kind filter '%d'", v.Int())
				}
			}

			for _, v := range filter.Authors() {
				if err := e.loadEvents(ctx, query.Where(eventFieldPublicKey, "==", v.Hex()), events, filter); err != nil {
					return nil, errors.Wrapf(err, "error loading events for author filter '%s'", v.Hex())
				}
			}

			for tagName, tagValues := range filter.Tags() {
				for _, tagValue := range tagValues {
					eventIds, err := e.tagRepository.GetEventIds(ctx, tagName, tagValue, filter.Since(), filter.Until(), filter.Limit())
					if err != nil {
						return nil, errors.Wrapf(err, "error loading events for tag '%s'->'%s'", tagName.String(), tagValue)
					}

					for _, eventId := range eventIds {
						if _, ok := events[eventId.Hex()]; ok {
							continue
						}

						event, err := e.Get(ctx, eventId)
						if err != nil {
							return nil, errors.Wrap(err, "error getting an event")
						}

						events[event.Id().Hex()] = event
					}
				}
			}
		}
	}

	// it is in my opinion unclear how to apply the limit field with multiple filters
	var result []domain.Event
	for _, event := range events {
		if filters.Match(event) {
			result = append(result, event)
		}
	}
	return result, nil
}

func (e *EventRepository) loadEvents(ctx context.Context, query firestore.Query, events map[string]domain.Event, filter domain.Filter) error {
	if since := filter.Since(); since != nil {
		query = query.Where(eventFieldCreatedAt, ">", since)
	}

	if until := filter.Until(); until != nil {
		query = query.Where(eventFieldCreatedAt, "<", until)
	}

	if filter.Limit() > 0 {
		query = query.Limit(filter.Limit())
	}

	docs := query.Documents(ctx)
	for {
		doc, err := docs.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return errors.Wrap(err, "error getting next document")
		}

		fmt.Println("event doc", doc.CreateTime)

		event, err := e.readEvent(doc)
		if err != nil {
			return errors.Wrap(err, "error reading the event")
		}

		events[event.Id().Hex()] = event
	}

	return nil
}

func (e *EventRepository) readEvent(doc *firestore.DocumentSnapshot) (domain.Event, error) {
	data := make(map[string]any)
	if err := doc.DataTo(&data); err != nil {
		return domain.Event{}, errors.Wrap(err, "error reading document data")
	}

	event, err := domain.NewEventFromRaw(data[eventFieldRaw].([]byte))
	if err != nil {
		return domain.Event{}, errors.Wrap(err, "error creating the event")
	}

	return event, nil
}

func sendErr(ctx context.Context, ch chan<- app.EventOrError, err error) {
	select {
	case ch <- app.NewEventOrErrorWithError(err):
	case <-ctx.Done():
	}
}
