package app

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/planetary-social/nos-crossposting-service/service/domain/sessions"
)

type LoginOrRegister struct {
	twitterID    accounts.TwitterID
	accessToken  accounts.TwitterUserAccessToken
	accessSecret accounts.TwitterUserAccessSecret
}

func NewLoginOrRegister(
	twitterID accounts.TwitterID,
	accessToken accounts.TwitterUserAccessToken,
	accessSecret accounts.TwitterUserAccessSecret,
) LoginOrRegister {
	return LoginOrRegister{
		twitterID:    twitterID,
		accessToken:  accessToken,
		accessSecret: accessSecret,
	}
}

type LoginOrRegisterHandler struct {
	transactionProvider TransactionProvider
	accountIDGenerator  AccountIDGenerator
	sessionIDGenerator  SessionIDGenerator
	logger              logging.Logger
	metrics             Metrics
}

func NewLoginOrRegisterHandler(
	transactionProvider TransactionProvider,
	accountIDGenerator AccountIDGenerator,
	sessionIDGenerator SessionIDGenerator,
	logger logging.Logger,
	metrics Metrics,
) *LoginOrRegisterHandler {
	return &LoginOrRegisterHandler{
		transactionProvider: transactionProvider,
		accountIDGenerator:  accountIDGenerator,
		sessionIDGenerator:  sessionIDGenerator,
		logger:              logger.New("loginOrRegisterHandler"),
		metrics:             metrics,
	}
}

func (h *LoginOrRegisterHandler) Handle(ctx context.Context, cmd LoginOrRegister) (session *sessions.Session, err error) {
	defer h.metrics.StartApplicationCall("loginOrRegister").End(&err)

	var result *sessions.Session
	if err := h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		account, err := h.createOrGetAccount(adapters, cmd.twitterID)
		if err != nil {
			return errors.Wrap(err, "error getting or creating account")
		}

		sessionID, err := h.sessionIDGenerator.GenerateSessionID()
		if err != nil {
			return errors.Wrap(err, "error generating a new session id")
		}

		session, err := sessions.NewSession(sessionID, account.AccountID(), time.Now())
		if err != nil {
			return errors.Wrap(err, "error creating a new session")
		}

		userTokens := accounts.NewTwitterUserTokens(account.AccountID(), cmd.accessToken, cmd.accessSecret)

		if err := adapters.Sessions.Save(session); err != nil {
			return errors.Wrap(err, "error saving a session")
		}

		if err := adapters.UserTokens.Save(userTokens); err != nil {
			return errors.Wrap(err, "error saving user tokens")
		}

		result = session
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction error")
	}

	return result, nil
}

func (h *LoginOrRegisterHandler) createOrGetAccount(adapters Adapters, twitterID accounts.TwitterID) (*accounts.Account, error) {
	account, err := adapters.Accounts.GetByTwitterID(twitterID)
	if err != nil {
		if errors.Is(err, ErrAccountDoesNotExist) {
			account, err := h.createAccount(adapters, twitterID)
			if err != nil {
				return nil, errors.Wrap(err, "error creating the account")
			}
			return account, nil
		}
		return nil, errors.Wrap(err, "error getting the account")
	}
	return account, nil
}

func (h *LoginOrRegisterHandler) createAccount(adapters Adapters, twitterID accounts.TwitterID) (*accounts.Account, error) {
	accountID, err := h.accountIDGenerator.GenerateAccountID()
	if err != nil {
		return nil, errors.Wrap(err, "error creating an account id")
	}

	account, err := accounts.NewAccount(accountID, twitterID)
	if err != nil {
		return nil, errors.Wrap(err, "error creating a new account")
	}

	if err := adapters.Accounts.Save(account); err != nil {
		return nil, errors.Wrap(err, "error saving the new account")
	}

	return account, nil
}
