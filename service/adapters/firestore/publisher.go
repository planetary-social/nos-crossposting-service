package firestore

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/firestore"
	"github.com/ThreeDotsLabs/watermill"
	watermillfirestore "github.com/ThreeDotsLabs/watermill-firestore/pkg/firestore"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

const PubsubTopicEventSaved = "event_saved"

type Publisher struct {
	publisher *watermillfirestore.Publisher
	tx        *firestore.Transaction
}

func NewPublisher(
	publisher *watermillfirestore.Publisher,
	tx *firestore.Transaction,
) *Publisher {
	return &Publisher{
		publisher: publisher,
		tx:        tx,
	}
}

func (p Publisher) PublishEventSaved(ctx context.Context, id domain.EventId) error {
	payload := EventSavedPayload{EventId: id.Hex()}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "error marshaling the payload")
	}

	msg := message.NewMessage(watermill.NewULID(), payloadJSON)
	return p.publisher.PublishInTransaction(PubsubTopicEventSaved, p.tx, msg)
}

type EventSavedPayload struct {
	EventId string `json:"eventId"`
}
