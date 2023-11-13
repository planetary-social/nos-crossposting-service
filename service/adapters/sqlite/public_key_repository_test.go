package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/stretchr/testify/require"
)

func TestPublicKeyRepository_ItIsPossibleToRetrieveSavedData(t *testing.T) {
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

	createdAt := time.Now()
	publicKey := fixtures.SomePublicKey()

	linkedPublicKey, err := domain.NewLinkedPublicKey(accountID, publicKey, createdAt)
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err = adapters.PublicKeyRepository.Save(linkedPublicKey)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		results, err := adapters.PublicKeyRepository.List()
		require.NoError(t, err)

		require.Len(t, results, 1)
		require.Equal(t, linkedPublicKey.AccountID(), results[0].AccountID())
		require.Equal(t, linkedPublicKey.PublicKey(), results[0].PublicKey())
		require.Equal(t, linkedPublicKey.CreatedAt().Truncate(time.Second), results[0].CreatedAt().Truncate(time.Second))

		return nil
	})
	require.NoError(t, err)
}

func TestPublicKeyRepository_ListByPublicKeyReturnsOnlyRelevantData(t *testing.T) {
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

	createdAt := time.Now()
	publicKey1 := fixtures.SomePublicKey()
	publicKey2 := fixtures.SomePublicKey()

	linkedPublicKey1, err := domain.NewLinkedPublicKey(accountID, publicKey1, createdAt)
	require.NoError(t, err)

	linkedPublicKey2, err := domain.NewLinkedPublicKey(accountID, publicKey2, createdAt)
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err = adapters.PublicKeyRepository.Save(linkedPublicKey1)
		require.NoError(t, err)

		err = adapters.PublicKeyRepository.Save(linkedPublicKey2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		results, err := adapters.PublicKeyRepository.ListByPublicKey(publicKey1)
		require.NoError(t, err)

		require.Len(t, results, 1)
		require.Equal(t, linkedPublicKey1.AccountID(), results[0].AccountID())
		require.Equal(t, linkedPublicKey1.PublicKey(), results[0].PublicKey())
		require.Equal(t, linkedPublicKey1.CreatedAt().Truncate(time.Second), results[0].CreatedAt().Truncate(time.Second))

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		results, err := adapters.PublicKeyRepository.ListByAccountID(accountID)
		require.NoError(t, err)

		require.Len(t, results, 2)

		return nil
	})
	require.NoError(t, err)
}

func TestPublicKeyRepository_DeletingNonExistentKeyDoesNotReturnAnError(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	err := adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err := adapters.PublicKeyRepository.Delete(fixtures.SomeAccountID(), fixtures.SomePublicKey())
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)
}

func TestPublicKeyRepository_DeletingPublicKeysDeletesThem(t *testing.T) {
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

	createdAt := time.Now()
	publicKey1 := fixtures.SomePublicKey()
	publicKey2 := fixtures.SomePublicKey()

	linkedPublicKey1, err := domain.NewLinkedPublicKey(accountID, publicKey1, createdAt)
	require.NoError(t, err)

	linkedPublicKey2, err := domain.NewLinkedPublicKey(accountID, publicKey2, createdAt)
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err = adapters.PublicKeyRepository.Save(linkedPublicKey1)
		require.NoError(t, err)

		err = adapters.PublicKeyRepository.Save(linkedPublicKey2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		results, err := adapters.PublicKeyRepository.ListByAccountID(accountID)
		require.NoError(t, err)

		require.Len(t, results, 2)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err := adapters.PublicKeyRepository.Delete(accountID, publicKey1)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		results, err := adapters.PublicKeyRepository.ListByAccountID(accountID)
		require.NoError(t, err)

		require.Len(t, results, 1)

		return nil
	})
	require.NoError(t, err)
}

func TestPublicKeyRepository_CountCountsPublicKeys(t *testing.T) {
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

	createdAt := time.Now()
	publicKey1 := fixtures.SomePublicKey()
	publicKey2 := fixtures.SomePublicKey()

	linkedPublicKey1, err := domain.NewLinkedPublicKey(accountID, publicKey1, createdAt)
	require.NoError(t, err)

	linkedPublicKey2, err := domain.NewLinkedPublicKey(accountID, publicKey2, createdAt)
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		n, err := adapters.PublicKeyRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 0, n)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err = adapters.PublicKeyRepository.Save(linkedPublicKey1)
		require.NoError(t, err)

		err = adapters.PublicKeyRepository.Save(linkedPublicKey2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		n, err := adapters.PublicKeyRepository.Count()
		require.NoError(t, err)
		require.Equal(t, 2, n)

		return nil
	})
	require.NoError(t, err)
}
