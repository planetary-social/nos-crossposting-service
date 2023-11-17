package app

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

const (
	dropEventsIfNotPostedFor = 7 * 24 * time.Hour
)

var (
	whenEventWasAddedToTweetCreatedEvent = time.Date(2023, time.November, 17, 0, 0, 0, 0, time.UTC)
)

type SendTweet struct {
	accountID accounts.AccountID
	tweet     domain.Tweet
	event     *domain.Event
}

func NewSendTweet(accountID accounts.AccountID, tweet domain.Tweet, event *domain.Event) SendTweet {
	return SendTweet{
		accountID: accountID,
		tweet:     tweet,
		event:     event,
	}
}

func (s SendTweet) AccountID() accounts.AccountID {
	return s.accountID
}

func (s SendTweet) Tweet() domain.Tweet {
	return s.tweet
}

func (s SendTweet) Event() *domain.Event {
	return s.event
}

type SendTweetHandler struct {
	transactionProvider TransactionProvider
	twitter             Twitter
	currentTimeProvider CurrentTimeProvider
	logger              logging.Logger
	metrics             Metrics
}

func NewSendTweetHandler(
	transactionProvider TransactionProvider,
	twitter Twitter,
	currentTimeProvider CurrentTimeProvider,
	logger logging.Logger,
	metrics Metrics,
) *SendTweetHandler {
	return &SendTweetHandler{
		transactionProvider: transactionProvider,
		twitter:             twitter,
		currentTimeProvider: currentTimeProvider,
		logger:              logger.New("sendTweetHandler"),
		metrics:             metrics,
	}
}

func (h *SendTweetHandler) Handle(ctx context.Context, cmd SendTweet) (err error) {
	defer h.metrics.StartApplicationCall("sendTweet").End(&err)

	h.logger.
		Debug().
		WithField("accountID", cmd.accountID).
		WithField("tweet", cmd.tweet.Text()).
		Message("attempting to post a tweet")

	if cmd.event != nil {
		dropEventIfPostedBefore := h.currentTimeProvider.GetCurrentTime().Add(-dropEventsIfNotPostedFor)
		if cmd.event.CreatedAt().Before(dropEventIfPostedBefore) {
			return nil
		}
	} else {
		dropEventIfItIsNilAndCurrentTimeIsAfter := whenEventWasAddedToTweetCreatedEvent.Add(dropEventsIfNotPostedFor)
		if h.currentTimeProvider.GetCurrentTime().After(dropEventIfItIsNilAndCurrentTimeIsAfter) {
			return nil
		}
	}

	var userTokens *accounts.TwitterUserTokens
	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		tmp, err := adapters.UserTokens.Get(cmd.accountID)
		if err != nil {
			return errors.Wrap(err, "error getting user tokens")
		}

		userTokens = tmp
		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction error")
	}

	if err := h.twitter.PostTweet(ctx, userTokens.AccessToken(), userTokens.AccessSecret(), cmd.tweet); err != nil {
		return errors.Wrap(err, "error posting to twitter")
	}

	return nil
}
