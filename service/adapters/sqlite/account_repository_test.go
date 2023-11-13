package sqlite_test

import (
	"context"
	"testing"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/stretchr/testify/require"
)

func TestAccountRepository_GetByAccountIDReturnsPredefinedErrorWhenDataIsNotAvailable(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	err := adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		_, err := adapters.AccountRepository.GetByAccountID(fixtures.SomeAccountID())
		require.ErrorIs(t, err, app.ErrAccountDoesNotExist)
		return nil
	})
	require.NoError(t, err)
}

func TestAccountRepository_GetByTwitterIDReturnsPredefinedErrorWhenDataIsNotAvailable(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	err := adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		_, err := adapters.AccountRepository.GetByTwitterID(fixtures.SomeTwitterID())
		require.ErrorIs(t, err, app.ErrAccountDoesNotExist)
		return nil
	})
	require.NoError(t, err)
}

func TestAccountRepository_ItIsPossibleToRetrieveSavedData(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	accountID := fixtures.SomeAccountID()
	twitterID := fixtures.SomeTwitterID()

	account, err := accounts.NewAccount(accountID, twitterID)
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err = adapters.AccountRepository.Save(account)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		retrievedAccount, err := adapters.AccountRepository.GetByTwitterID(twitterID)
		require.NoError(t, err)
		require.Equal(t, account, retrievedAccount)

		return nil
	})
	require.NoError(t, err)
}

func TestAccountRepository_CountReturnsNumberOfAccounts(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	err := adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		n, err := adapters.AccountRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 0, n)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		accountID := fixtures.SomeAccountID()
		twitterID := fixtures.SomeTwitterID()

		account, err := accounts.NewAccount(accountID, twitterID)
		require.NoError(t, err)
		err = adapters.AccountRepository.Save(account)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		n, err := adapters.AccountRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 1, n)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		accountID := fixtures.SomeAccountID()
		twitterID := fixtures.SomeTwitterID()

		account, err := accounts.NewAccount(accountID, twitterID)
		require.NoError(t, err)
		err = adapters.AccountRepository.Save(account)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		n, err := adapters.AccountRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 2, n)

		return nil
	})
	require.NoError(t, err)
}
