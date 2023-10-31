package sqlite

import (
	"database/sql"
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/oklog/ulid/v2"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

const TweetCreatedTopic = "tweet_created"

type Publisher struct {
	pubsub *PubSub
	tx     *sql.Tx
}

func NewPublisher(pubsub *PubSub, tx *sql.Tx) *Publisher {
	return &Publisher{pubsub: pubsub, tx: tx}
}

func (p *Publisher) PublishTweetCreated(accountID accounts.AccountID, tweet domain.Tweet) error {
	transport := TweetCreatedEventTransport{
		AccountID: accountID.String(),
		Tweet: TweetTransport{
			Text: tweet.Text(),
		},
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
}

type TweetTransport struct {
	Text string `json:"text"`
}
