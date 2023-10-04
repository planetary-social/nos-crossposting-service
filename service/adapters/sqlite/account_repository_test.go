package sqlite_test

import (
	"context"
	"testing"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite/tests"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/stretchr/testify/require"
)

func TestAccountRepository_GetByAccountIDReturnsPredefinedErrorWhenDataIsNotAvailable(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := tests.NewTestAdapters(ctx, t)

	err := adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		_, err := adapters.AccountRepository.GetByAccountID(fixtures.SomeAccountID())
		require.ErrorIs(t, err, app.ErrAccountDoesNotExist)
		return nil
	})
	require.NoError(t, err)
}

func TestAccountRepository_GetByTwittterIDReturnsPredefinedErrorWhenDataIsNotAvailable(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := tests.NewTestAdapters(ctx, t)

	err := adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		_, err := adapters.AccountRepository.GetByTwitterID(fixtures.SomeTwitterID())
		require.ErrorIs(t, err, app.ErrAccountDoesNotExist)
		return nil
	})
	require.NoError(t, err)
}
