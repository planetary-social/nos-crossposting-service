package sqlite

import (
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	watermillsql "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

const TweetCreatedTopic = "tweet_created"

type Publisher struct {
	watermillPublisher *watermillsql.Publisher
}

func NewPublisher(watermillPublisher *watermillsql.Publisher) *Publisher {
	return &Publisher{watermillPublisher: watermillPublisher}
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

	msg := message.NewMessage(watermill.NewULID(), payload)
	return p.watermillPublisher.Publish(TweetCreatedTopic, msg)
}

type TweetCreatedEventTransport struct {
	AccountID string         `json:"accountID"`
	Tweet     TweetTransport `json:"tweet"`
}

type TweetTransport struct {
	Text string `json:"text"`
}
