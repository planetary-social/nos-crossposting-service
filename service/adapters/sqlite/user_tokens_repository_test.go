package sqlite_test

import (
	"context"
	"testing"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/stretchr/testify/require"
)

func TestUserTokensRepository_ItIsPossibleToSaveTokensAndThenReadThem(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	accountID := fixtures.SomeAccountID()
	twitterID := fixtures.SomeTwitterID()

	account, err := accounts.NewAccount(accountID, twitterID)
	require.NoError(t, err)

	accessToken, err := accounts.NewTwitterUserAccessToken(fixtures.SomeString())
	require.NoError(t, err)

	accessSecret, err := accounts.NewTwitterUserAccessSecret(fixtures.SomeString())
	require.NoError(t, err)

	userTokens := accounts.NewTwitterUserTokens(accountID, accessToken, accessSecret)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err = adapters.AccountRepository.Save(account)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err = adapters.UserTokensRepository.Save(userTokens)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		tokens, err := adapters.UserTokensRepository.Get(accountID)
		require.NoError(t, err)

		require.Equal(t, accountID, tokens.AccountID())
		require.Equal(t, accessToken, tokens.AccessToken())
		require.Equal(t, accessSecret, tokens.AccessSecret())

		return nil
	})
	require.NoError(t, err)
}
