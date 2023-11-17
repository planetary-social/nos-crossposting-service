package mocks

import (
	"context"

	"github.com/planetary-social/nos-crossposting-service/service/app"
)

type TransactionProvider struct {
	adapters app.Adapters
}

func NewTransactionProvider(adapters app.Adapters) *TransactionProvider {
	return &TransactionProvider{adapters: adapters}
}

func (t *TransactionProvider) Transact(ctx context.Context, f func(context.Context, app.Adapters) error) error {
	return f(ctx, t.adapters)
}
