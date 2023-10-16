package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriber_TweetCreatedQueueLength(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	n, err := adapters.Subscriber.TweetCreatedQueueLength(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, n)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err := adapters.Publisher.PublishTweetCreated(fixtures.SomeAccountID(), domain.NewTweet(fixtures.SomeString()))
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err := adapters.Publisher.PublishTweetCreated(fixtures.SomeAccountID(), domain.NewTweet(fixtures.SomeString()))
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	n, err = adapters.Subscriber.TweetCreatedQueueLength(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	go func() {
		ch, err := adapters.Subscriber.SubscribeToTweetCreated(ctx)
		require.NoError(t, err)

		for msg := range ch {
			msg.Ack()
		}
	}()

	require.EventuallyWithT(t, func(t *assert.CollectT) {
		n, err := adapters.Subscriber.TweetCreatedQueueLength(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, n)
	}, 5*time.Second, 100*time.Millisecond)
}
