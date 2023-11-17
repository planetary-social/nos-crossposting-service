package mocks

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type Twitter struct {
	PostTweetCalls []PostTweetCall
}

func NewTwitter() *Twitter {
	return &Twitter{}
}

func (t *Twitter) PostTweet(ctx context.Context, userAccessToken accounts.TwitterUserAccessToken, userAccessSecret accounts.TwitterUserAccessSecret, tweet domain.Tweet) error {
	t.PostTweetCalls = append(t.PostTweetCalls, PostTweetCall{
		UserAccessToken:  userAccessToken,
		UserAccessSecret: userAccessSecret,
		Tweet:            tweet,
	})
	return nil
}

func (t *Twitter) GetAccountDetails(ctx context.Context, userAccessToken accounts.TwitterUserAccessToken, userAccessSecret accounts.TwitterUserAccessSecret) (app.TwitterAccountDetails, error) {
	return app.TwitterAccountDetails{}, errors.New("not implemented")
}

type PostTweetCall struct {
	UserAccessToken  accounts.TwitterUserAccessToken
	UserAccessSecret accounts.TwitterUserAccessSecret
	Tweet            domain.Tweet
}
