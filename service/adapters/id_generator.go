package adapters

import (
	"github.com/oklog/ulid/v2"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/planetary-social/nos-crossposting-service/service/domain/sessions"
)

type IDGenerator struct {
}

func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

func (I IDGenerator) GenerateSessionID() (sessions.SessionID, error) {
	return sessions.NewSessionID(ulid.Make().String())
}

func (I IDGenerator) GenerateAccountID() (accounts.AccountID, error) {
	return accounts.NewAccountID(ulid.Make().String())
}
