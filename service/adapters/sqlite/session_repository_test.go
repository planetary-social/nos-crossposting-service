package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain/sessions"
	"github.com/stretchr/testify/require"
)

func TestSessionRepository_GetReturnsPredefinedErrorWhenDataIsNotAvailable(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	err := adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		_, err := adapters.SessionRepository.Get(fixtures.SomeSessionID())
		require.ErrorIs(t, err, app.ErrSessionDoesNotExist)
		return nil
	})
	require.NoError(t, err)
}

func TestSessionRepository_ItIsPossibleToRetrieveSavedData(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	sessionID := fixtures.SomeSessionID()
	accountID := fixtures.SomeAccountID()
	createdAt := time.Now()

	session, err := sessions.NewSession(sessionID, accountID, createdAt)
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err = adapters.SessionRepository.Save(session)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		retrievedSession, err := adapters.SessionRepository.Get(sessionID)
		require.NoError(t, err)
		require.Equal(t, session.SessionID(), retrievedSession.SessionID())
		require.Equal(t, session.AccountID(), retrievedSession.AccountID())
		require.Equal(t, session.CreatedAt().UTC().Truncate(time.Second), retrievedSession.CreatedAt().UTC().Truncate(time.Second))

		return nil
	})
	require.NoError(t, err)
}

func TestSessionRepository_DeletingNonexistentSessionReturnsNoError(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	err := adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err := adapters.SessionRepository.Delete(fixtures.SomeSessionID())
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)
}

func TestSessionRepository_DeletingSessionDeletesSession(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	sessionID := fixtures.SomeSessionID()
	accountID := fixtures.SomeAccountID()
	createdAt := time.Now()

	session, err := sessions.NewSession(sessionID, accountID, createdAt)
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err = adapters.SessionRepository.Save(session)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		_, err := adapters.SessionRepository.Get(sessionID)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		err := adapters.SessionRepository.Delete(sessionID)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	err = adapters.TransactionProvider.Transact(ctx, func(ctx context.Context, adapters sqlite.TestAdapters) error {
		_, err := adapters.SessionRepository.Get(sessionID)
		require.ErrorIs(t, err, app.ErrSessionDoesNotExist)
		return nil
	})
	require.NoError(t, err)
}
