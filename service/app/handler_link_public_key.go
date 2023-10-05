package app

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type LinkPublicKey struct {
	accountID accounts.AccountID
	publicKey domain.PublicKey
}

func NewLinkPublicKey(accountID accounts.AccountID, publicKey domain.PublicKey) LinkPublicKey {
	return LinkPublicKey{accountID: accountID, publicKey: publicKey}
}

type LinkPublicKeyHandler struct {
	transactionProvider TransactionProvider
	logger              logging.Logger
	metrics             Metrics
}

func NewLinkPublicKeyHandler(
	transactionProvider TransactionProvider,
	logger logging.Logger,
	metrics Metrics,
) *LinkPublicKeyHandler {
	return &LinkPublicKeyHandler{
		transactionProvider: transactionProvider,
		logger:              logger.New("linkPublicKeyHandler"),
		metrics:             metrics,
	}
}

func (h *LinkPublicKeyHandler) Handle(ctx context.Context, cmd LinkPublicKey) (err error) {
	defer h.metrics.StartApplicationCall("linkPublicKey").End(&err)

	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		linkedPublicKey, err := domain.NewLinkedPublicKey(cmd.accountID, cmd.publicKey, time.Now())
		if err != nil {
			return errors.Wrap(err, "error creating a linked public key")
		}

		if err := adapters.PublicKeys.Save(linkedPublicKey); err != nil {
			return errors.Wrap(err, "error saving the linked public key")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction error")
	}

	return nil
}
