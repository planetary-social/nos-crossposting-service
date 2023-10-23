package app

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type GetTwitterAccountDetails struct {
	accountID accounts.AccountID
}

func NewGetTwitterAccountDetails(accountID accounts.AccountID) GetTwitterAccountDetails {
	return GetTwitterAccountDetails{accountID: accountID}
}

type GetTwitterAccountDetailsHandler struct {
	transactionProvider TransactionProvider
	twitter             Twitter
	logger              logging.Logger
	metrics             Metrics
}

func NewGetTwitterAccountDetailsHandler(
	transactionProvider TransactionProvider,
	twitter Twitter,
	logger logging.Logger,
	metrics Metrics,
) *GetTwitterAccountDetailsHandler {
	return &GetTwitterAccountDetailsHandler{
		transactionProvider: transactionProvider,
		twitter:             twitter,
		logger:              logger.New("getTwitterAccountDetailsHandler"),
		metrics:             metrics,
	}
}

func (h *GetTwitterAccountDetailsHandler) Handle(ctx context.Context, cmd GetTwitterAccountDetails) (result TwitterAccountDetails, err error) {
	defer h.metrics.StartApplicationCall("getTwitterAccountDetails").End(&err)

	var userTokens *accounts.TwitterUserTokens
	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		tmp, err := adapters.UserTokens.Get(cmd.accountID)
		if err != nil {
			return errors.Wrap(err, "error getting user tokens")
		}

		userTokens = tmp
		return nil
	}); err != nil {
		return TwitterAccountDetails{}, errors.Wrap(err, "transaction error")
	}

	twitterAccountDetails, err := h.twitter.GetAccountDetails(ctx, userTokens.AccessToken(), userTokens.AccessSecret())
	if err != nil {
		return TwitterAccountDetails{}, errors.Wrap(err, "error getting twitter account details")
	}

	return twitterAccountDetails, nil
}
