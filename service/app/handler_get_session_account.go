package app

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/planetary-social/nos-crossposting-service/service/domain/sessions"
)

type GetSessionAccount struct {
	sessionID sessions.SessionID
}

func NewGetSessionAccount(sessionID sessions.SessionID) GetSessionAccount {
	return GetSessionAccount{sessionID: sessionID}
}

type GetSessionAccountHandler struct {
	transactionProvider TransactionProvider
	logger              logging.Logger
	metrics             Metrics
}

func NewGetSessionAccountHandler(
	transactionProvider TransactionProvider,
	logger logging.Logger,
	metrics Metrics,
) *GetSessionAccountHandler {
	return &GetSessionAccountHandler{
		transactionProvider: transactionProvider,
		logger:              logger.New("getSessionAccountHandler"),
		metrics:             metrics,
	}
}

func (h *GetSessionAccountHandler) Handle(ctx context.Context, cmd GetSessionAccount) (result *accounts.Account, err error) {
	defer h.metrics.StartApplicationCall("getSessionAccount").End(&err)

	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		session, err := adapters.Sessions.Get(cmd.sessionID)
		if err != nil {
			return errors.Wrap(err, "error getting a session")
		}

		account, err := adapters.Accounts.GetByAccountID(session.AccountID())
		if err != nil {
			return errors.Wrap(err, "error getting an account")
		}

		result = account
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction error")
	}

	return result, nil
}
