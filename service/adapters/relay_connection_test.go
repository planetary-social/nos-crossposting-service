package adapters_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters"
	"github.com/stretchr/testify/require"
)

func TestDialErrorIs(t *testing.T) {
	err := adapters.NewDialError(fixtures.SomeError())
	require.ErrorIs(t, err, adapters.DialError{})
	require.ErrorIs(t, err, &adapters.DialError{})
	require.ErrorIs(t, errors.Wrap(err, "wrapped"), adapters.DialError{})
	require.ErrorIs(t, errors.Wrap(err, "wrapped"), &adapters.DialError{})
}

func TestReadMessageErrorIs(t *testing.T) {
	err := adapters.NewReadMessageError(fixtures.SomeError())
	require.ErrorIs(t, err, adapters.ReadMessageError{})
	require.ErrorIs(t, err, &adapters.ReadMessageError{})
	require.ErrorIs(t, errors.Wrap(err, "wrapped"), adapters.ReadMessageError{})
	require.ErrorIs(t, errors.Wrap(err, "wrapped"), &adapters.ReadMessageError{})
}
