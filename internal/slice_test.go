package internal_test

import (
	"testing"

	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/stretchr/testify/require"
)

func TestName(t *testing.T) {
	require.Equal(t,
		[][]int(nil),
		internal.BatchesFromSlice(
			[]int{},
			3,
		),
	)

	require.Equal(t,
		[][]int{
			{
				1, 2, 3,
			},
		},
		internal.BatchesFromSlice(
			[]int{
				1, 2, 3,
			},
			5,
		),
	)

	require.Equal(t,
		[][]int{
			{
				1, 2, 3,
			},
			{
				4, 5, 6,
			},
			{
				7, 8, 9,
			},
			{
				10,
			},
		},
		internal.BatchesFromSlice(
			[]int{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
			},
			3,
		),
	)
}
