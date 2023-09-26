package app

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

type GetPublicKeysHandler struct {
	transactionProvider TransactionProvider
	metrics             Metrics
}

func NewGetPublicKeysHandler(
	transactionProvider TransactionProvider,
	metrics Metrics,
) *GetPublicKeysHandler {
	return &GetPublicKeysHandler{
		transactionProvider: transactionProvider,
		metrics:             metrics,
	}
}

func (h *GetPublicKeysHandler) Handle(ctx context.Context, relay domain.RelayAddress) (keys []domain.PublicKey, err error) {
	defer h.metrics.StartApplicationCall("getPublicKeys").End(&err)

	var result []domain.PublicKey
	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		tmp, err := adapters.Relays.GetPublicKeys(ctx, relay, time.Now().Add(-getPublicKeysYoungerThan))
		if err != nil {
			return errors.Wrap(err, "error getting relays")
		}
		result = tmp
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction error")
	}
	return result, nil
}
