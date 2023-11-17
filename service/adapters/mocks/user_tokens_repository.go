package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type UserTokensRepository struct {
	mockedUserTokens map[accounts.AccountID]*accounts.TwitterUserTokens
}

func NewUserTokensRepository() (*UserTokensRepository, error) {
	return &UserTokensRepository{
		mockedUserTokens: make(map[accounts.AccountID]*accounts.TwitterUserTokens),
	}, nil
}

func (m *UserTokensRepository) Save(userTokens *accounts.TwitterUserTokens) error {
	return errors.New("not implemented")
}

func (m *UserTokensRepository) Get(id accounts.AccountID) (*accounts.TwitterUserTokens, error) {
	v, ok := m.mockedUserTokens[id]
	if !ok {
		return nil, errors.New("user tokens not mocked")
	}
	return v, nil
}

func (m *UserTokensRepository) MockUserTokens(tokens *accounts.TwitterUserTokens) {
	m.mockedUserTokens[tokens.AccountID()] = tokens
}
