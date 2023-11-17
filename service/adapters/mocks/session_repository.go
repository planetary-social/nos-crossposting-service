package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain/sessions"
)

type SessionRepository struct {
}

func NewSessionRepository() (*SessionRepository, error) {
	return &SessionRepository{}, nil
}

func (m *SessionRepository) Get(id sessions.SessionID) (*sessions.Session, error) {
	return nil, errors.New("not implemented")
}

func (m *SessionRepository) Save(session *sessions.Session) error {
	return errors.New("not implemented")
}

func (m *SessionRepository) Delete(id sessions.SessionID) error {
	return errors.New("not implemented")
}
