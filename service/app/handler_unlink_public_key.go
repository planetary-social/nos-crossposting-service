package app

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type UnlinkPublicKey struct {
	accountID accounts.AccountID
	publicKey domain.PublicKey
}

func NewUnlinkPublicKey(accountID accounts.AccountID, publicKey domain.PublicKey) UnlinkPublicKey {
	return UnlinkPublicKey{accountID: accountID, publicKey: publicKey}
}

type UnlinkPublicKeyHandler struct {
	transactionProvider TransactionProvider
	logger              logging.Logger
	metrics             Metrics
}

func NewUnlinkPublicKeyHandler(
	transactionProvider TransactionProvider,
	logger logging.Logger,
	metrics Metrics,
) *UnlinkPublicKeyHandler {
	return &UnlinkPublicKeyHandler{
		transactionProvider: transactionProvider,
		logger:              logger.New("unlinkPublicKeyHandler"),
		metrics:             metrics,
	}
}

func (h *UnlinkPublicKeyHandler) Handle(ctx context.Context, cmd UnlinkPublicKey) (err error) {
	defer h.metrics.StartApplicationCall("unlinkPublicKey").End(&err)

	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		if err := adapters.PublicKeys.Delete(cmd.accountID, cmd.publicKey); err != nil {
			return errors.Wrap(err, "error deleting the linked publicm key")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction error")
	}

	return nil
}
