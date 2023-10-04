package http

import (
	"net/http"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain/sessions"
)

const sessionIDCookieName = "sessionID"

var ErrNoSessionID = errors.New("no session id found in cookies")

func GetSessionIDFromCookie(r *http.Request) (sessions.SessionID, error) {
	cookie, err := r.Cookie(sessionIDCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return sessions.SessionID{}, ErrNoSessionID
		}
		return sessions.SessionID{}, errors.Wrap(err, "error getting the session cookie")
	}
	return sessions.NewSessionID(cookie.Value)
}

func SetSessionIDToCookie(rw http.ResponseWriter, sessionID sessions.SessionID) {
	cookie := &http.Cookie{
		Name:  sessionIDCookieName,
		Value: sessionID.String(),
	}
	http.SetCookie(rw, cookie)
}
