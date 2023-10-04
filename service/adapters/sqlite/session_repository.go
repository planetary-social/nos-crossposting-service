package sqlite

import (
	"database/sql"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/planetary-social/nos-crossposting-service/service/domain/sessions"
)

type SessionRepository struct {
	tx *sql.Tx
}

func NewSessionRepository(tx *sql.Tx) (*SessionRepository, error) {
	return &SessionRepository{
		tx: tx,
	}, nil
}

func (m *SessionRepository) Get(id sessions.SessionID) (*sessions.Session, error) {
	result := m.tx.QueryRow(`
SELECT session_id, account_id, created_at
FROM sessions
WHERE session_id=$1`,
		id.String(),
	)

	return m.readSession(result)
}

func (m *SessionRepository) Save(session *sessions.Session) error {
	_, err := m.tx.Exec(`
INSERT INTO sessions(session_id, account_id, created_at)
VALUES($1, $2, $3)
ON CONFLICT(session_id) DO UPDATE SET
  account_id=excluded.account_id,
  created_at=excluded.created_at`,
		session.SessionID().String(),
		session.AccountID().String(),
		session.CreatedAt().Unix(),
	)
	if err != nil {
		return errors.Wrap(err, "error executing the insert query")
	}

	return nil
}

func (m *SessionRepository) readSession(result *sql.Row) (*sessions.Session, error) {
	var sessionIDTmp string
	var accountIDTmp string
	var createdAtTmp int64

	if err := result.Scan(&sessionIDTmp, &accountIDTmp, &createdAtTmp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, app.ErrSessionDoesNotExist
		}
		return nil, errors.Wrap(err, "error reading the row")
	}

	sessionID, err := sessions.NewSessionID(sessionIDTmp)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the session id")
	}

	accountID, err := accounts.NewAccountID(accountIDTmp)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the account id")
	}

	createdAt := time.Unix(createdAtTmp, 0)

	return sessions.NewSession(sessionID, accountID, createdAt)
}
