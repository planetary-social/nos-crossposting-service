package sqlite_test

import (
	"testing"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/migrations"
	"github.com/stretchr/testify/require"
)

func TestMigrationsStorage_LoadStateReturnsCorrectErrorWhenStateIsNotAvailable(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	_, err := adapters.MigrationsStorage.LoadState(fixtures.SomeString())
	require.ErrorIs(t, err, migrations.ErrStateNotFound)
}

func TestMigrationsStorage_LoadStateReturnsSavedState(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	name := fixtures.SomeString()
	state := migrations.State{
		fixtures.SomeString(): fixtures.SomeString(),
	}

	err := adapters.MigrationsStorage.SaveState(name, state)
	require.NoError(t, err)

	loadedState, err := adapters.MigrationsStorage.LoadState(name)
	require.NoError(t, err)
	require.Equal(t, state, loadedState)
}

func TestMigrationsStorage_SavingStateTwiceOverwritesPreviousState(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	name := fixtures.SomeString()
	state1 := migrations.State{
		fixtures.SomeString(): fixtures.SomeString(),
	}
	state2 := migrations.State{
		fixtures.SomeString(): fixtures.SomeString(),
	}

	err := adapters.MigrationsStorage.SaveState(name, state1)
	require.NoError(t, err)

	err = adapters.MigrationsStorage.SaveState(name, state2)
	require.NoError(t, err)

	loadedState, err := adapters.MigrationsStorage.LoadState(name)
	require.NoError(t, err)
	require.Equal(t, state2, loadedState)
}

func TestMigrationsStorage_LoadStatusReturnsCorrectErrorWhenStatusIsNotAvailable(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	_, err := adapters.MigrationsStorage.LoadStatus(fixtures.SomeString())
	require.ErrorIs(t, err, migrations.ErrStatusNotFound)
}

func TestMigrationsStorage_LoadStatusReturnsSavedStatus(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	name := fixtures.SomeString()
	status := migrations.StatusFinished

	err := adapters.MigrationsStorage.SaveStatus(name, status)
	require.NoError(t, err)

	loadedStatus, err := adapters.MigrationsStorage.LoadStatus(name)
	require.NoError(t, err)
	require.Equal(t, status, loadedStatus)
}

func TestMigrationsStorage_SavingStatusTwiceOverwritesPreviousStatus(t *testing.T) {
	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	name := fixtures.SomeString()
	status1 := migrations.StatusFinished
	status2 := migrations.StatusFinished

	err := adapters.MigrationsStorage.SaveStatus(name, status1)
	require.NoError(t, err)

	err = adapters.MigrationsStorage.SaveStatus(name, status2)
	require.NoError(t, err)

	loadedStatus, err := adapters.MigrationsStorage.LoadStatus(name)
	require.NoError(t, err)
	require.Equal(t, status2, loadedStatus)
}
