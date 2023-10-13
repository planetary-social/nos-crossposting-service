package twitter

import (
	"context"

	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type NoopTwitter struct {
	logger logging.Logger
}

func NewNoopTwitter(logger logging.Logger) *NoopTwitter {
	return &NoopTwitter{
		logger: logger.New("noopTwitter"),
	}
}

func (t *NoopTwitter) PostTweet(
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
