package sqlite_test

import (
	"context"
	"testing"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/stretchr/testify/require"
)

func TestProcessedEventRepository_WasProcessedReturnsFalseIfEventWasNotProcessed(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	err := adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		wasProcessed, err := adapters.ProcessedEventRepository.WasProcessed(fixtures.SomeEventID(), fixtures.SomeTwitterID())
		require.NoError(t, err)
		require.False(t, wasProcessed)

		return nil
	})
	require.NoError(t, err)
}

func TestProcessedEventRepository_WasProcessedReturnsTrueIfEventWasProcessed(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	eventID := fixtures.SomeEventID()
	twitterID := fixtures.SomeTwitterID()

	err := adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err := adapters.ProcessedEventRepository.Save(eventID, twitterID)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		wasProcessed, err := adapters.ProcessedEventRepository.WasProcessed(eventID, twitterID)
		require.NoError(t, err)
		require.True(t, wasProcessed)

		return nil
	})
	require.NoError(t, err)
}

func TestProcessedEventRepository_CallingWasProcessedTwiceReturnsNoErrors(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	eventID := fixtures.SomeEventID()
	twitterID := fixtures.SomeTwitterID()

	err := adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err := adapters.ProcessedEventRepository.Save(eventID, twitterID)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err := adapters.ProcessedEventRepository.Save(eventID, twitterID)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)
}
