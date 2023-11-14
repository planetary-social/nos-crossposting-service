package twitter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/boreq/errors"
	"github.com/g8rswimmer/go-twitter/v2"
	oauth1 "github.com/klaidas/go-oauth1"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

const (
	apiLimitWindow         = 15 * time.Minute
	apiLimitCreateTweet    = 50 // docs claim 200 but it doesn't seem true at all
	apiLimitGetUserDetails = 75
)

type Twitter struct {
	conf    config.Config
	logger  logging.Logger
	metrics app.Metrics
	limiter *Limiter
}

func NewTwitter(
	conf config.Config,
	logger logging.Logger,
	metrics app.Metrics,
) *Twitter {
	return &Twitter{
		conf:    conf,
		limiter: NewLimiter(),
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
		t.conf,
		userAccessToken,
		userAccessSecret,
		nil,
	)

	client := &twitter.Client{
		Authorizer: authorizer,
		Client:     http.DefaultClient,
		Host:       "https://api.twitter.com",
	}

	if err := t.limiter.Limit(
		fmt.Sprintf("create-tweet-%s", userAccessToken),
		apiLimitCreateTweet,
		apiLimitWindow,
	); err != nil {
		return errors.Wrap(err, "limiter error")
	}

	response, err := client.CreateTweet(ctx, twitter.CreateTweetRequest{
		Text: tweet.Text(),
	})
	err = t.convertError(err)
	t.metrics.ReportCallingTwitterAPIToPostATweet(err)
	if err != nil {
		t.logError(err)
		return errors.Wrap(err, "error calling create tweet")
	}

	t.logger.Debug().
		WithField("tweetID", response.Tweet.ID).
		Message("posted a tweet")

	return nil
}

func (t *Twitter) GetAccountDetails(
	ctx context.Context,
	userAccessToken accounts.TwitterUserAccessToken,
	userAccessSecret accounts.TwitterUserAccessSecret,
) (app.TwitterAccountDetails, error) {
	authorizer := newUserAuthorizer(
		t.conf,
		userAccessToken,
		userAccessSecret,
		map[string]string{
			"user.fields": "username,name,profile_image_url",
		},
	)

	client := &twitter.Client{
		Authorizer: authorizer,
		Client:     http.DefaultClient,
		Host:       "https://api.twitter.com",
	}

	if err := t.limiter.Limit(
		fmt.Sprintf("user-lookup-%s", userAccessToken),
		apiLimitGetUserDetails,
		apiLimitWindow,
	); err != nil {
		return app.TwitterAccountDetails{}, errors.Wrap(err, "limiter error")
	}

	result, err := client.UserLookup(ctx, []string{"me"}, twitter.UserLookupOpts{
		UserFields: []twitter.UserField{
			twitter.UserFieldUserName,
			twitter.UserFieldName,
			twitter.UserFieldProfileImageURL,
		},
	})
	err = t.convertError(err)
	t.metrics.ReportCallingTwitterAPIToGetAUser(err)
	if err != nil {
		t.logError(err)
		return app.TwitterAccountDetails{}, errors.Wrap(err, "error looking up the user")
	}

	if len(result.Raw.Users) != 1 {
		return app.TwitterAccountDetails{}, errors.Wrap(err, "expected 1 user")
	}

	user := result.Raw.Users[0]
	return app.NewTwitterAccountDetails(user.Name, user.UserName, user.ProfileImageURL)
}

func (t *Twitter) logError(err error) {
	var errorResponse *twitter.ErrorResponse
	if errors.As(err, &errorResponse) {
		l := t.logger.Error().
			WithField("statusCode", errorResponse.StatusCode).
			WithField("title", errorResponse.Title).
			WithField("detail", errorResponse.Detail).
			WithField("type", errorResponse.Type)

		for i, error := range errorResponse.Errors {
			l = l.
				WithField(fmt.Sprintf("err.%d.parameters", i), error.Parameters).
				WithField(fmt.Sprintf("err.%d.message", i), error.Message)
		}

		if errorResponse.RateLimit != nil {
			l = l.WithField("rateLimit.limit", errorResponse.RateLimit.Limit).
				WithField("rateLimit.reset", errorResponse.RateLimit.Reset).
				WithField("rateLimit.remaining", errorResponse.RateLimit.Remaining)
		}

		l.Message("received an error response from twitter")
	}
}

func (t *Twitter) convertError(err error) error {
	if err == nil {
		return nil
	}

	var errorResponse *twitter.ErrorResponse
	if errors.As(err, &errorResponse) {
		return NewTwitterError(errorResponse)
	}

	return err
}

type userAuthorizer struct {
	conf             config.Config
	userAccessToken  accounts.TwitterUserAccessToken
	userAccessSecret accounts.TwitterUserAccessSecret
	params           map[string]string
}

func newUserAuthorizer(
	conf config.Config,
	userAccessToken accounts.TwitterUserAccessToken,
	userAccessSecret accounts.TwitterUserAccessSecret,
	params map[string]string,
) *userAuthorizer {
	if params == nil {
		params = make(map[string]string)
	}

	return &userAuthorizer{
		conf:             conf,
		userAccessToken:  userAccessToken,
		userAccessSecret: userAccessSecret,
		params:           params,
	}
}

func (a *userAuthorizer) Add(req *http.Request) {
	auth := oauth1.OAuth1{
		ConsumerKey:    a.conf.TwitterKey(),
		ConsumerSecret: a.conf.TwitterKeySecret(),
		AccessToken:    a.userAccessToken.String(),
		AccessSecret:   a.userAccessSecret.String(),
	}

	authHeader := auth.BuildOAuth1Header(req.Method, req.URL.String(), a.params)
	req.Header.Set("Authorization", authHeader)
}

type TwitterError struct {
	underlying *twitter.ErrorResponse
}

func NewTwitterError(underlying *twitter.ErrorResponse) TwitterError {
	return TwitterError{underlying: underlying}
}

func (t TwitterError) Error() string {
	return fmt.Sprintf("twitter error: %s", t.underlying)
}

func (t TwitterError) Unwrap() error {
	return t.underlying
}

func (t TwitterError) Description() string {
	return t.underlying.Title
}

func (t TwitterError) Is(target error) bool {
	_, ok1 := target.(TwitterError)
	_, ok2 := target.(*TwitterError)
	return ok1 || ok2
}
