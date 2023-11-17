package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/stretchr/testify/require"
)

func TestPublisher_ItIsPossibleToPublishEvents(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	accountID := fixtures.SomeAccountID()
	tweet := domain.NewTweet("some tweet")
	twitterID := fixtures.SomeTwitterID()
	createdAt := time.Now()
	event := fixtures.SomeEvent()

	tweetCreatedEvent := app.NewTweetCreatedEvent(accountID, tweet, createdAt, event)

	account, err := accounts.NewAccount(accountID, twitterID)
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err = adapters.AccountRepository.Save(account)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err = adapters.Publisher.PublishTweetCreated(tweetCreatedEvent)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)
}
