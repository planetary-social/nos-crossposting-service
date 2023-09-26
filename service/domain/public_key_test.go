package domain_test

import (
	"strings"
	"testing"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/stretchr/testify/require"
)

func TestPublicKey_IsCaseInsensitive(t *testing.T) {
	hex := fixtures.SomeHexBytesOfLen(32)
	hexLower := strings.ToLower(hex)
	hexUpper := strings.ToUpper(hex)

	require.NotEqual(t, hexLower, hexUpper)

	a, err := domain.NewPublicKeyFromHex(hexLower)
	require.NoError(t, err)

	b, err := domain.NewPublicKeyFromHex(hexUpper)
	require.NoError(t, err)

	require.Equal(t, a, b)
}
