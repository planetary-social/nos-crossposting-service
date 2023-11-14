package twitter_test

import (
	"testing"

	"github.com/boreq/errors"
	twitterlib "github.com/g8rswimmer/go-twitter/v2"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/twitter"
	"github.com/stretchr/testify/require"
)

func TestErrorIs(t *testing.T) {
	someError := &twitterlib.ErrorResponse{}
	err := twitter.NewTwitterError(someError)
	require.ErrorIs(t, err, twitter.TwitterError{})
	require.ErrorIs(t, err, &twitter.TwitterError{})
	require.ErrorIs(t, errors.Wrap(err, "wrapped"), twitter.TwitterError{})
	require.ErrorIs(t, errors.Wrap(err, "wrapped"), &twitter.TwitterError{})
}
