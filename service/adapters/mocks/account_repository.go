package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type AccountRepository struct {
}

func NewAccountRepository() (*AccountRepository, error) {
	return &AccountRepository{}, nil
}

func (m *AccountRepository) GetByTwitterID(twitterID accounts.TwitterID) (*accounts.Account, error) {
	return nil, errors.New("not implemented")
}

func (m *AccountRepository) GetByAccountID(accountID accounts.AccountID) (*accounts.Account, error) {
	return nil, errors.New("not implemented")
}

func (m *AccountRepository) Save(account *accounts.Account) error {
	return errors.New("not implemented")
}

func (m *AccountRepository) Count() (int, error) {
	return 0, errors.New("not implemented")
}
