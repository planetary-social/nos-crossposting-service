package sqlitepubsub

import (
	"context"
	"encoding/json"

	watermillsql "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type SendTweetHandler interface {
	Handle(ctx context.Context, cmd app.SendTweet) (err error)
}

type TweetCreatedEventSubscriber struct {
	handler             SendTweetHandler
	watermillSubscriber *watermillsql.Subscriber
	logger              logging.Logger
}

func NewTweetCreatedEventSubscriber(
	handler SendTweetHandler,
	watermillSubscriber *watermillsql.Subscriber,
	logger logging.Logger,
) *TweetCreatedEventSubscriber {
	return &TweetCreatedEventSubscriber{
		handler:             handler,
		watermillSubscriber: watermillSubscriber,
		logger:              logger.New("tweetCreatedEventSubscriber"),
	}
}

func (s *TweetCreatedEventSubscriber) Run(ctx context.Context) error {
	ch, err := s.watermillSubscriber.Subscribe(ctx, sqlite.TweetCreatedTopic)
	if err != nil {
		return errors.Wrap(err, "error calling subscribe")
	}

	for msg := range ch {
		if err := s.handleMessage(ctx, msg); err != nil {
			s.logger.Error().WithError(err).Message("error handling a message")
			msg.Nack()
		} else {
			msg.Ack()
		}
	}

	return errors.New("channel closed")
}

func (s *TweetCreatedEventSubscriber) handleMessage(ctx context.Context, msg *message.Message) error {
	var transport sqlite.TweetCreatedEventTransport
	if err := json.Unmarshal(msg.Payload, &transport); err != nil {
		return errors.Wrap(err, "error unmarshaling")
	}

	accountID, err := accounts.NewAccountID(transport.AccountID)
	if err != nil {
		return errors.Wrap(err, "error creating an account id")
	}

	tweet := domain.NewTweet(transport.Tweet.Text)
	cmd := app.NewSendTweet(accountID, tweet)

	if err := s.handler.Handle(ctx, cmd); err != nil {
		return errors.Wrap(err, "error calling the handler")
	}

	return nil
}
