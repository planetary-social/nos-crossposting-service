package sessions

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type Session struct {
	accountID accounts.AccountID
	sessionID SessionID
}

func NewSession(accountID accounts.AccountID, sessionID SessionID) *Session {
	return &Session{accountID: accountID, sessionID: sessionID}
}

func (s Session) AccountID() accounts.AccountID {
	return s.accountID
}

func (s Session) SessionID() SessionID {
	return s.sessionID
}

type SessionID struct {
	id string
}

func NewSessionID(id string) (SessionID, error) {
	if id == "" {
		return SessionID{}, errors.New("session id can't be an empty string")
	}
	return SessionID{id: id}, nil
}

func (i SessionID) String() string {
	return i.id
}
