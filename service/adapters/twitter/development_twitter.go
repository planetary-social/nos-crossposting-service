package twitter

import (
	"context"

	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type DevelopmentTwitter struct {
	logger logging.Logger
}

func NewDevelopmentTwitter(logger logging.Logger) *DevelopmentTwitter {
	return &DevelopmentTwitter{
		logger: logger.New("noopOnWriteTwitter"),
	}
}

func (t *DevelopmentTwitter) PostTweet(
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

func (t *DevelopmentTwitter) GetAccountDetails(
	ctx context.Context,
	userAccessToken accounts.TwitterUserAccessToken,
	userAccessSecret accounts.TwitterUserAccessSecret,
) (app.TwitterAccountDetails, error) {
	// it is too easy to hit API limits using a free API key during development
	return app.NewTwitterAccountDetails(
		"Fake Display Name",
		"fakeusername",
		"https://pbs.twimg.com/profile_images/1544326468490170368/VCPwpDkL_normal.jpg",
	)
}
