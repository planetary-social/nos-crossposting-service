package sqlitepubsub

import (
	"context"
	"encoding/json"

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

type SqliteSubscriber interface {
	SubscribeToTweetCreated(ctx context.Context) <-chan *sqlite.ReceivedMessage
}

type TweetCreatedEventSubscriber struct {
	handler    SendTweetHandler
	subscriber SqliteSubscriber
	logger     logging.Logger
}

func NewTweetCreatedEventSubscriber(
	handler SendTweetHandler,
	subscriber SqliteSubscriber,
	logger logging.Logger,
) *TweetCreatedEventSubscriber {
	return &TweetCreatedEventSubscriber{
		handler:    handler,
		subscriber: subscriber,
		logger:     logger.New("tweetCreatedEventSubscriber"),
	}
}
func (s *TweetCreatedEventSubscriber) Run(ctx context.Context) error {
	for msg := range s.subscriber.SubscribeToTweetCreated(ctx) {
		if err := s.handleMessage(ctx, msg); err != nil {
			s.logger.Error().WithError(err).Message("error handling a message")
			if err := msg.Nack(); err != nil {
				return errors.Wrap(err, "error nacking a message")
			}
		} else {
			if err := msg.Ack(); err != nil {
				return errors.Wrap(err, "error acking a message")
			}
		}
	}

	return errors.New("channel closed")
}

func (s *TweetCreatedEventSubscriber) handleMessage(ctx context.Context, msg *sqlite.ReceivedMessage) error {
	var transport sqlite.TweetCreatedEventTransport
	if err := json.Unmarshal(msg.Payload(), &transport); err != nil {
		return errors.Wrap(err, "error unmarshaling")
	}

	accountID, err := accounts.NewAccountID(transport.AccountID)
	if err != nil {
		return errors.Wrap(err, "error creating an account id")
	}

	tweet := domain.NewTweet(transport.Tweet.Text)

	event, err := domain.NewEventFromRaw(transport.Event)
	if err != nil {
		return errors.Wrap(err, "error loading the event")
	}

	cmd := app.NewSendTweet(accountID, tweet, event)

	if err := s.handler.Handle(ctx, cmd); err != nil {
		return errors.Wrap(err, "error calling the handler")
	}

	return nil
}
