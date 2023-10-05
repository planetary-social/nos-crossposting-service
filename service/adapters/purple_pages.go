package adapters

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

type PurplePages struct {
}

func NewPurplePages() *PurplePages {
	return &PurplePages{}
}

func (p PurplePages) GetRelays(ctx context.Context, publicKey domain.PublicKey) ([]domain.RelayAddress, error) {
	// todo
	return nil, errors.New("not implemented")
}
