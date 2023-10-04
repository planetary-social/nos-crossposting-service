package sessions

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type Session struct {
	sessionID SessionID
	accountID accounts.AccountID
	createdAt time.Time
}

func NewSession(sessionID SessionID, accountID accounts.AccountID, createdAt time.Time) (*Session, error) {
	if createdAt.IsZero() {
		return nil, errors.New("zero value of created at")
	}
	return &Session{
		sessionID: sessionID,
		accountID: accountID,
		createdAt: createdAt,
	}, nil
}

func (s Session) SessionID() SessionID {
	return s.sessionID
}

func (s Session) AccountID() accounts.AccountID {
	return s.accountID
}

func (s Session) CreatedAt() time.Time {
	return s.createdAt
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
