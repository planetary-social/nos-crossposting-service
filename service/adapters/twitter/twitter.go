package twitter

import (
	"context"
	"net/http"

	"github.com/boreq/errors"
	"github.com/g8rswimmer/go-twitter/v2"
	oauth1 "github.com/klaidas/go-oauth1"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type Twitter struct {
	config  config.Config
	logger  logging.Logger
	metrics app.Metrics
}

func NewTwitter(
	config config.Config,
	logger logging.Logger,
	metrics app.Metrics,
) *Twitter {
	return &Twitter{
		config:  config,
		logger:  logger.New("twitter"),
		metrics: metrics,
	}
}

func (t *Twitter) PostTweet(
	ctx context.Context,
	userAccessToken accounts.TwitterUserAccessToken,
	userAccessSecret accounts.TwitterUserAccessSecret,
	tweet domain.Tweet,
) error {
	authorizer := newUserAuthorizer(
		t.config,
		userAccessToken,
		userAccessSecret,
		tweet,
	)

	client := &twitter.Client{
		Authorizer: authorizer,
		Client:     http.DefaultClient,
		Host:       "https://api.twitter.com",
	}

	response, err := client.CreateTweet(ctx, twitter.CreateTweetRequest{
		Text: tweet.Text(),
	})
	t.metrics.ReportCallingTwitterAPIToPostATweet(err)
	if err != nil {
		var errorResponse *twitter.ErrorResponse
		if errors.As(err, &errorResponse) {
			t.logger.Error().
				WithField("statusCode", errorResponse.StatusCode).
				WithField("title", errorResponse.Title).
				WithField("detail", errorResponse.Detail).
				WithField("type", errorResponse.Type).
				Message("received an error response from twitter")
		}
		return errors.Wrap(err, "error calling create tweet")
	}

	t.logger.Debug().
		WithField("tweetID", response.Tweet.ID).
		Message("posted a tweet")

	return nil
}

type userAuthorizer struct {
	logger           logging.Logger
	config           config.Config
	userAccessToken  accounts.TwitterUserAccessToken
	userAccessSecret accounts.TwitterUserAccessSecret
	tweet            domain.Tweet
}

func newUserAuthorizer(
	config config.Config,
	userAccessToken accounts.TwitterUserAccessToken,
	userAccessSecret accounts.TwitterUserAccessSecret,
	tweet domain.Tweet,
) *userAuthorizer {
	return &userAuthorizer{
		config:           config,
		userAccessToken:  userAccessToken,
		userAccessSecret: userAccessSecret,
		tweet:            tweet,
	}
}

func (a *userAuthorizer) Add(req *http.Request) {
	auth := oauth1.OAuth1{
		ConsumerKey:    a.config.TwitterKey(),
		ConsumerSecret: a.config.TwitterKeySecret(),
		AccessToken:    a.userAccessToken.String(),
		AccessSecret:   a.userAccessSecret.String(),
	}

	authHeader := auth.BuildOAuth1Header(req.Method, req.URL.String(), map[string]string{})
	req.Header.Set("Authorization", authHeader)
}
