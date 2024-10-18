package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type PublicKeyRepository struct {
}

func NewPublicKeyRepository() (*PublicKeyRepository, error) {
	return &PublicKeyRepository{}, nil
}

func (m *PublicKeyRepository) Save(linkedPublicKey *domain.LinkedPublicKey) error {
	return errors.New("not implemented")
}

func (m *PublicKeyRepository) Delete(accountID accounts.AccountID, publicKey domain.PublicKey) error {
	return errors.New("not implemented")
}

func (m *PublicKeyRepository) DeleteByPublicKey(publicKey domain.PublicKey) error {
	return errors.New("not implemented")
}

func (m *PublicKeyRepository) List() ([]*domain.LinkedPublicKey, error) {
	return nil, errors.New("not implemented")
}

func (m *PublicKeyRepository) ListByPublicKey(publicKey domain.PublicKey) ([]*domain.LinkedPublicKey, error) {
	return nil, errors.New("not implemented")
}

func (m *PublicKeyRepository) ListByAccountID(accountID accounts.AccountID) ([]*domain.LinkedPublicKey, error) {
	return nil, errors.New("not implemented")
}

func (m *PublicKeyRepository) Count() (int, error) {
	return 0, errors.New("not implemented")
}
