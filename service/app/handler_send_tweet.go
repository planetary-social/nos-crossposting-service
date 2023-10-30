package app

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type SendTweet struct {
	accountID accounts.AccountID
	tweet     domain.Tweet
}

func NewSendTweet(accountID accounts.AccountID, tweet domain.Tweet) SendTweet {
	return SendTweet{accountID: accountID, tweet: tweet}
}

type SendTweetHandler struct {
	transactionProvider TransactionProvider
	twitter             Twitter
	logger              logging.Logger
	metrics             Metrics
}

func NewSendTweetHandler(
	transactionProvider TransactionProvider,
	twitter Twitter,
	logger logging.Logger,
	metrics Metrics,
) *SendTweetHandler {
	return &SendTweetHandler{
		transactionProvider: transactionProvider,
		twitter:             twitter,
		logger:              logger.New("sendTweetHandler"),
		metrics:             metrics,
	}
}

func (h *SendTweetHandler) Handle(ctx context.Context, cmd SendTweet) (err error) {
	defer h.metrics.StartApplicationCall("sendTweet").End(&err)

	h.logger.Debug().
		WithField("accountID", cmd.accountID).
		WithField("tweet", cmd.tweet.Text()).
		Message("attempting to post a tweet")

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
