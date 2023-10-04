package memory

import (
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type MemoryAccountRepository struct {
	accounts map[string]*accounts.Account
}

func NewMemoryAccountRepository() *MemoryAccountRepository {
	return &MemoryAccountRepository{
		accounts: make(map[string]*accounts.Account),
	}
}

func (m *MemoryAccountRepository) GetByTwitterID(twitterID accounts.TwitterID) (*accounts.Account, error) {
	for _, account := range m.accounts {
		if account.TwitterID() == twitterID {
			return account, nil
		}
	}
	return nil, app.ErrAccountDoesNotExist
}

func (m *MemoryAccountRepository) GetByAccountID(accountID accounts.AccountID) (*accounts.Account, error) {
	for _, account := range m.accounts {
		if account.AccountID() == accountID {
			return account, nil
		}
	}
	return nil, app.ErrAccountDoesNotExist
}

func (m *MemoryAccountRepository) Save(account *accounts.Account) error {
	m.accounts[account.AccountID().String()] = account
	return nil
}
