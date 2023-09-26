package app

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/notifications"
)

type GetNotificationsHandler struct {
	transactionProvider TransactionProvider
	metrics             Metrics
}

func NewGetNotificationsHandler(
	transactionProvider TransactionProvider,
	metrics Metrics,
) *GetNotificationsHandler {
	return &GetNotificationsHandler{
		transactionProvider: transactionProvider,
		metrics:             metrics,
	}
}

func (h *GetNotificationsHandler) Handle(ctx context.Context, id domain.EventId) (result []notifications.Notification, err error) {
	defer h.metrics.StartApplicationCall("getNotifications").End(&err)

	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		tmp, err := adapters.Events.GetNotifications(ctx, id)
		if err != nil {
			return errors.Wrap(err, "error getting notifications")
		}
		result = tmp
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction error")
	}
	return result, nil
}
