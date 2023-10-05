package domain

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type LinkedPublicKey struct {
	accountID accounts.AccountID
	publicKey PublicKey
	createdAt time.Time
}

func NewLinkedPublicKey(accountID accounts.AccountID, publicKey PublicKey, createdAt time.Time) (*LinkedPublicKey, error) {
	if createdAt.IsZero() {
		return nil, errors.New("created at can't be zero")
	}
	return &LinkedPublicKey{
		accountID: accountID,
		publicKey: publicKey,
		createdAt: createdAt,
	}, nil
}

func (l LinkedPublicKey) AccountID() accounts.AccountID {
	return l.accountID
}

func (l LinkedPublicKey) PublicKey() PublicKey {
	return l.publicKey
}

func (l LinkedPublicKey) CreatedAt() time.Time {
	return l.createdAt
}
