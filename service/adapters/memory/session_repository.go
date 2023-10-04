package memory

import (
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain/sessions"
)

type MemorySessionRepository struct {
	sessions map[string]*sessions.Session
}

func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{
		sessions: make(map[string]*sessions.Session),
	}
}

func (m MemorySessionRepository) Get(id sessions.SessionID) (*sessions.Session, error) {
	for _, session := range m.sessions {
		if session.SessionID() == id {
			return session, nil
		}
	}
	return nil, app.ErrSessionDoesNotExist
}

func (m MemorySessionRepository) Save(session *sessions.Session) error {
	m.sessions[session.SessionID().String()] = session
	return nil
}
