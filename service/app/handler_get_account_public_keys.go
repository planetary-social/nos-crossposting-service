package app

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type GetAccountPublicKeys struct {
	accountID accounts.AccountID
}

func NewGetAccountPublicKeys(accountID accounts.AccountID) GetAccountPublicKeys {
	return GetAccountPublicKeys{accountID: accountID}
}

type GetAccountPublicKeysHandler struct {
	transactionProvider TransactionProvider
	logger              logging.Logger
	metrics             Metrics
}

func NewGetAccountPublicKeysHandler(
	transactionProvider TransactionProvider,
	logger logging.Logger,
	metrics Metrics,
) *GetAccountPublicKeysHandler {
	return &GetAccountPublicKeysHandler{
		transactionProvider: transactionProvider,
		logger:              logger.New("getAccountPublicKeys"),
		metrics:             metrics,
	}
}

func (h *GetAccountPublicKeysHandler) Handle(ctx context.Context, cmd GetAccountPublicKeys) (result []*domain.LinkedPublicKey, err error) {
	defer h.metrics.StartApplicationCall("getAccountPublicKeys").End(&err)

	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		publicKeys, err := adapters.PublicKeys.ListByAccountID(cmd.accountID)
		if err != nil {
			return errors.Wrap(err, "error getting a session")
		}

		result = publicKeys
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction error")
	}

	return result, nil
}
