package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRelayAddress(t *testing.T) {
	testCases := []struct {
		In  string
		Out string
	}{
		{
			In:  "wss://nos.social",
			Out: "wss://nos.social",
		},
		{
			In:  "wss://nos.social/",
			Out: "wss://nos.social",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.In, func(t *testing.T) {
			address, err := NormalizeRelayAddress(MustNewRelayAddress(testCase.In))
			require.NoError(t, err)
			require.Equal(t, testCase.Out, address.String())
		})
	}
}
