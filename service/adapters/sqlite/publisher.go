package sqlite

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/boreq/errors"
	"github.com/oklog/ulid/v2"
	"github.com/planetary-social/nos-crossposting-service/service/app"
)

const TweetCreatedTopic = "tweet_created"

type Publisher struct {
	pubsub *PubSub
	tx     *sql.Tx
}

func NewPublisher(pubsub *PubSub, tx *sql.Tx) *Publisher {
	return &Publisher{pubsub: pubsub, tx: tx}
}

func (p *Publisher) PublishTweetCreated(event app.TweetCreatedEvent) error {
	transport := TweetCreatedEventTransport{
		AccountID: event.AccountID().String(),
		Tweet: TweetTransport{
			Text: event.Tweet().Text(),
		},
		Event:     event.Event().Raw(),
		CreatedAt: event.CreatedAt(),
	}

	payload, err := json.Marshal(transport)
	if err != nil {
		return errors.Wrap(err, "error marshaling the transport type")
	}

	msg, err := NewMessage(ulid.Make().String(), payload)
	if err != nil {
		return errors.Wrap(err, "error creating a message")
	}

	return p.pubsub.PublishTx(p.tx, TweetCreatedTopic, msg)
}

type TweetCreatedEventTransport struct {
	AccountID string         `json:"accountID"`
	Tweet     TweetTransport `json:"tweet"`
	Event     []byte         `json:"event"`
	CreatedAt time.Time      `json:"createdAt"`
}

type TweetTransport struct {
	Text string `json:"text"`
}
