package sqlite_test

import (
	"context"
	"testing"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/stretchr/testify/require"
)

func TestSubscriber_TweetCreatedAnalysis(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	err := adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		for i := 0; i < 10; i++ {
			accountID := fixtures.SomeAccountID()

			for j := 0; j <= i; j++ {
				err := adapters.Publisher.PublishTweetCreated(accountID, domain.NewTweet(fixtures.SomeString()))
				require.NoError(t, err)
			}
		}

		return nil
	})
	require.NoError(t, err)

	analysis, err := adapters.Subscriber.TweetCreatedAnalysis(ctx)
	require.NoError(t, err)
	require.Equal(t, 10, len(analysis.TweetsPerAccountID))
	for _, count := range analysis.TweetsPerAccountID {
		require.NotZero(t, count)
	}
}
