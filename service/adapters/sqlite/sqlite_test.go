package sqlite_test

import (
	"context"
	"testing"

	"github.com/planetary-social/nos-crossposting-service/cmd/crossposting-service/di"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/stretchr/testify/require"
)

func NewTestAdapters(ctx context.Context, tb testing.TB) sqlite.TestedItems {
	adapters, f, err := di.BuildTestAdapters(ctx, tb)
	require.NoError(tb, err)

	tb.Cleanup(f)

	err = adapters.Migrations.Execute(ctx)
	require.NoError(tb, err)

	return adapters
}
