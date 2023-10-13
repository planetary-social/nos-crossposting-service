package twitter

import (
	"context"

	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type TwitterMock struct {
	logger logging.Logger
}

func NewTwitterMock(logger logging.Logger) *TwitterMock {
	return &TwitterMock{
		logger: logger.New("twitterMock"),
	}
}

func (t *TwitterMock) PostTweet(
	ctx context.Context,
	userAccessToken accounts.TwitterUserAccessToken,
	userAccessSecret accounts.TwitterUserAccessSecret,
	tweet domain.Tweet,
) error {
	t.logger.Debug().
		WithField("text", tweet.Text()).
		Message("triggered posting a tweet in mock Twitter adapter")
	return nil
}
