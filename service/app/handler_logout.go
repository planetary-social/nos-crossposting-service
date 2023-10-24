package app

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain/sessions"
)

type Logout struct {
	sessionID sessions.SessionID
}

func NewLogout(sessionID sessions.SessionID) Logout {
	return Logout{sessionID: sessionID}
}

type LogoutHandler struct {
	transactionProvider TransactionProvider
	logger              logging.Logger
	metrics             Metrics
}

func NewLogoutHandler(
	transactionProvider TransactionProvider,
	logger logging.Logger,
	metrics Metrics,
) *LogoutHandler {
	return &LogoutHandler{
		transactionProvider: transactionProvider,
		logger:              logger.New("logoutHandler"),
		metrics:             metrics,
	}
}

func (h *LogoutHandler) Handle(ctx context.Context, cmd Logout) (err error) {
	defer h.metrics.StartApplicationCall("logout").End(&err)

	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		return adapters.Sessions.Delete(cmd.sessionID)
	}); err != nil {
		return errors.Wrap(err, "transaction error")
	}

	return nil
}
