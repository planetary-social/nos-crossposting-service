package app

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
)

type UpdateMetricsHandler struct {
	transactionProvider TransactionProvider
	subscriber          Subscriber
	logger              logging.Logger
	metrics             Metrics
}

func NewUpdateMetricsHandler(
	transactionProvider TransactionProvider,
	subscriber Subscriber,
	logger logging.Logger,
	metrics Metrics,
) *UpdateMetricsHandler {
	return &UpdateMetricsHandler{
		transactionProvider: transactionProvider,
		subscriber:          subscriber,
		logger:              logger.New("updateMetricsHandler"),
		metrics:             metrics,
	}
}

func (h *UpdateMetricsHandler) Handle(ctx context.Context) (err error) {
	defer h.metrics.StartApplicationCall("updateMetrics").End(&err)

	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		n, err := adapters.Accounts.Count()
		if err != nil {
			return errors.Wrap(err, "error counting accounts")
		}
		h.metrics.ReportNumberOfAccounts(n)

		n, err = adapters.PublicKeys.Count()
		if err != nil {
			return errors.Wrap(err, "error counting linked public keys")
		}
		h.metrics.ReportNumberOfLinkedPublicKeys(n)

		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction error")
	}

	n, err := h.subscriber.TweetCreatedQueueLength(ctx)
	if err != nil {
		return errors.Wrap(err, "error reading queue length")
	}
	h.metrics.ReportSubscriptionQueueLength("tweet_created", n)

	analysis, err := h.subscriber.TweetCreatedAnalysis(ctx)
	if err != nil {
		return errors.Wrap(err, "error reading queue analysis")
	}
	h.metrics.ReportTweetCreatedCountPerAccount(analysis.TweetsPerAccountID)

	return nil
}
