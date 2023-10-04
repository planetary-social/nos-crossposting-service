package http

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/boreq/errors"
	"github.com/dghubble/gologin/v2/twitter"
	"github.com/dghubble/oauth1"
	twitterOAuth1 "github.com/dghubble/oauth1/twitter"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

//go:embed templates/*
var templatesFS embed.FS

var t = template.Must(template.ParseFS(templatesFS, "templates/*.tmpl"))

type Server struct {
	config config.Config
	app    app.Application
	logger logging.Logger
}

func NewServer(
	config config.Config,
	app app.Application,
	logger logging.Logger,
) Server {
	return Server{
		config: config,
		app:    app,
		logger: logger.New("server"),
	}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := s.createMux()

	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(ctx, "tcp", s.config.ListenAddress())
	if err != nil {
		return errors.Wrap(err, "error listening")
	}

	go func() {
		<-ctx.Done()
		if err := listener.Close(); err != nil {
			fmt.Println("error closing listener:", err)
		}
	}()

	return http.Serve(listener, mux)
}

func (s *Server) createMux() *http.ServeMux {
	config := &oauth1.Config{
		ConsumerKey:    s.config.TwitterKey(),
		ConsumerSecret: s.config.TwitterKeySecret(),
		CallbackURL:    "http://localhost:8008/callback", // todo config?
		Endpoint:       twitterOAuth1.AuthorizeEndpoint,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.serveIndex)
	mux.Handle("/login", twitter.LoginHandler(config, nil))
	mux.Handle("/callback", twitter.CallbackHandler(config, s.issueSession(), nil))
	return mux
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	account, err := s.getAccountFromRequest(r)
	if err != nil {
		s.renderError(w, err)
		return
	}

	data := s.templateDataFromAccount(account)

	if err := t.ExecuteTemplate(w, "index.tmpl", data); err != nil {
		s.logger.Error().WithError(err).Message("error rendering index")
	}
}

func (s *Server) renderError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("error"))
}

func (s *Server) issueSession() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		twitterUser, err := twitter.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		twitterID := accounts.NewTwitterID(twitterUser.ID)
		cmd := app.NewLoginOrRegister(twitterID)

		session, err := s.app.LoginOrRegister.Handle(req.Context(), cmd)
		if err != nil {
			s.renderError(w, err)
			return
		}

		SetSessionIDToCookie(w, session.SessionID())

		s.logger.Debug().
			WithField("twitterID", twitterID.Int64()).
			WithField("accountID", session.AccountID().String()).
			WithField("sessionID", session.SessionID().String()).
			Message("issuing a session")

		http.Redirect(w, req, "/", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

func (s *Server) getAccountFromRequest(r *http.Request) (*accounts.Account, error) {
	sessionID, err := GetSessionIDFromCookie(r)
	if err != nil {
		if errors.Is(err, ErrNoSessionID) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "error getting the session id from cookie")
	}

	cmd := app.NewGetSessionAccount(sessionID)

	account, err := s.app.GetSessionAccount.Handle(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, app.ErrSessionDoesNotExist) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "error getting the account")
	}

	return account, nil
}

func (s *Server) templateDataFromAccount(account *accounts.Account) map[string]any {
	if account == nil {
		return map[string]any{
			"account": nil,
		}
	}

	return map[string]any{
		"account": accountTransport{
			AccountID: account.AccountID().String(),
			TwitterID: account.TwitterID().Int64(),
		},
	}
}

type accountTransport struct {
	AccountID string
	TwitterID int64
}
