package twitter

import (
	"context"

	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type NoopOnWriteTwitter struct {
	twitter *Twitter
	logger  logging.Logger
}

func NewNoopOnWriteTwitter(twitter *Twitter, logger logging.Logger) *NoopOnWriteTwitter {
	return &NoopOnWriteTwitter{
		twitter: twitter,
		logger:  logger.New("noopOnWriteTwitter"),
	}
}

func (t *NoopOnWriteTwitter) PostTweet(
	ctx context.Context,
	userAccessToken accounts.TwitterUserAccessToken,
	userAccessSecret accounts.TwitterUserAccessSecret,
	tweet domain.Tweet,
) error {
	t.logger.Debug().
		WithField("text", tweet.Text()).
		Message("triggered posting a tweet in a noop Twitter adapter")
	return nil
}

func (t *NoopOnWriteTwitter) GetAccountDetails(
	ctx context.Context,
	userAccessToken accounts.TwitterUserAccessToken,
	userAccessSecret accounts.TwitterUserAccessSecret,
) (app.TwitterAccountDetails, error) {
	return t.twitter.GetAccountDetails(ctx, userAccessToken, userAccessSecret)
}
