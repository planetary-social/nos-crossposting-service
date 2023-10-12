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
	return &TwitterMock{logger: logger}
}

func (t *TwitterMock) PostTweet(
	ctx context.Context,
	userAccessToken accounts.TwitterUserAccessToken,
	userAccessSecret accounts.TwitterUserAccessSecret,
	tweet domain.Tweet,
) error {
	t.logger.Debug().Message("posted a mock tweet")
	return nil
}
