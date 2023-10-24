package http

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/boreq/errors"
	"github.com/boreq/rest"
	oauth12 "github.com/dghubble/gologin/v2/oauth1"
	"github.com/dghubble/gologin/v2/twitter"
	"github.com/dghubble/oauth1"
	twitterOAuth1 "github.com/dghubble/oauth1/twitter"
	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/planetary-social/nos-crossposting-service/service/ports/http/frontend"
)

const loginCallbackPath = `/login-callback`

type Server struct {
	conf               config.Config
	app                app.Application
	logger             logging.Logger
	frontendFileSystem *frontend.FrontendFileSystem
}

func NewServer(
	conf config.Config,
	app app.Application,
	logger logging.Logger,
	frontendFileSystem *frontend.FrontendFileSystem,
) Server {
	return Server{
		conf:               conf,
		app:                app,
		logger:             logger.New("server"),
		frontendFileSystem: frontendFileSystem,
	}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := s.createMux()

	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(ctx, "tcp", s.conf.ListenAddress())
	if err != nil {
		return errors.Wrap(err, "error listening")
	}

	s.logger.
		Debug().
		WithField("address", s.conf.ListenAddress()).
		Message("started the listener")

	go func() {
		<-ctx.Done()
		if err := listener.Close(); err != nil {
			s.logger.Error().WithError(err).Message("error closing listener")
		}
	}()

	return http.Serve(listener, mux)
}

func (s *Server) createMux() *http.ServeMux {
	config := &oauth1.Config{
		ConsumerKey:    s.conf.TwitterKey(),
		ConsumerSecret: s.conf.TwitterKeySecret(),
		CallbackURL:    s.twitterLoginCallbackURL(),
		Endpoint:       twitterOAuth1.AuthorizeEndpoint,
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(s.frontendFileSystem))
	mux.Handle("/login", twitter.LoginHandler(config, nil))
	mux.HandleFunc("/api/current-user", rest.Wrap(s.apiCurrentUser))
	mux.HandleFunc("/api/public-keys", rest.Wrap(s.apiPublicKeys))
	mux.Handle(loginCallbackPath, twitter.CallbackHandler(config, s.issueSession(), nil))

	return mux
}

func (s *Server) twitterLoginCallbackURL() string {
	base := strings.TrimRight(s.conf.PublicFacingAddress(), "/")
	return base + loginCallbackPath
}

func (s *Server) issueSession() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if err := s.issueSessionErr(w, req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, req, "/", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

func (s *Server) issueSessionErr(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()

	twitterUser, err := twitter.UserFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "error getting twitter user from context")
	}

	accessTokenString, accessSecretString, err := oauth12.AccessTokenFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "error getting access token from context")
	}

	twitterID := accounts.NewTwitterID(twitterUser.ID)

	accessToken, err := accounts.NewTwitterUserAccessToken(accessTokenString)
	if err != nil {
		return errors.Wrap(err, "error creating user access token")
	}

	accessSecret, err := accounts.NewTwitterUserAccessSecret(accessSecretString)
	if err != nil {
		return errors.Wrap(err, "error creating user access secret")
	}

	cmd := app.NewLoginOrRegister(twitterID, accessToken, accessSecret)

	session, err := s.app.LoginOrRegister.Handle(req.Context(), cmd)
	if err != nil {
		return errors.Wrap(err, "error calling login or register handler")
	}

	SetSessionIDToCookie(w, session.SessionID())

	s.logger.Debug().
		WithField("twitterID", twitterID.Int64()).
		WithField("accountID", session.AccountID().String()).
		WithField("sessionID", session.SessionID().String()).
		Message("issuing a session")

	return nil
}

func (s *Server) apiCurrentUser(r *http.Request) rest.RestResponse {
	account, err := s.getAccountFromRequest(r)
	if err != nil {
		s.logger.Error().WithError(err).Message("error getting account from request")
		return rest.ErrInternalServerError
	}

	if account == nil {
		return rest.NewResponse(
			currentUserResponse{
				User: nil,
			},
		)
	}

	twitterAccountDetails, err := s.app.GetTwitterAccountDetails.Handle(r.Context(), app.NewGetTwitterAccountDetails(account.AccountID()))
	if err != nil {
		s.logger.Error().WithError(err).Message("error getting twitter account details")
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(
		currentUserResponse{
			User: internal.Pointer(newTransportUser(*account, twitterAccountDetails)),
		},
	)
}

func (s *Server) apiPublicKeys(r *http.Request) rest.RestResponse {
	switch r.Method {
	case http.MethodGet:
		return s.apiPublicKeysList(r)
	case http.MethodPost:
		return s.apiPublicKeysAdd(r)
	default:
		return rest.ErrMethodNotAllowed
	}
}

func (s *Server) apiPublicKeysList(r *http.Request) rest.RestResponse {
	ctx := r.Context()

	account, err := s.getAccountFromRequest(r)
	if err != nil {
		s.logger.Error().WithError(err).Message("error getting account from request")
		return rest.ErrInternalServerError
	}

	if account == nil {
		return rest.ErrUnauthorized
	}

	linkedPublicKeys, err := s.app.GetAccountPublicKeys.Handle(ctx, app.NewGetAccountPublicKeys(account.AccountID()))
	if err != nil {
		s.logger.Error().WithError(err).Message("error getting public keys")
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(
		publicKeysListResponse{
			PublicKeys: newTransportPublicKeys(linkedPublicKeys),
		},
	)
}

func (s *Server) apiPublicKeysAdd(r *http.Request) rest.RestResponse {
	ctx := r.Context()

	account, err := s.getAccountFromRequest(r)
	if err != nil {
		s.logger.Error().WithError(err).Message("error getting account from request")
		return rest.ErrInternalServerError
	}

	if account == nil {
		return rest.ErrUnauthorized
	}

	var t publicKeysAddRequest
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return rest.ErrBadRequest
	}

	publicKey, err := domain.NewPublicKeyFromNpub(t.Npub)
	if err != nil {
		return rest.ErrBadRequest
	}

	cmd := app.NewLinkPublicKey(account.AccountID(), publicKey)

	if err := s.app.LinkPublicKey.Handle(ctx, cmd); err != nil {
		s.logger.Error().WithError(err).Message("error adding a public key")
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(nil)
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

type currentUserResponse struct {
	User *transportUser `json:"user"`
}

type publicKeysListResponse struct {
	PublicKeys []transportPublicKey `json:"publicKeys"`
}

type publicKeysAddRequest struct {
	Npub string `json:"npub"`
}

type transportUser struct {
	AccountID              string `json:"accountID"`
	TwitterID              int64  `json:"twitterID"`
	TwitterName            string `json:"twitterName"`
	TwitterUsername        string `json:"twitterUsername"`
	TwitterProfileImageURL string `json:"twitterProfileImageURL"`
}

func newTransportUser(account accounts.Account, twitterAccountDetails app.TwitterAccountDetails) transportUser {
	return transportUser{
		AccountID:              account.AccountID().String(),
		TwitterID:              account.TwitterID().Int64(),
		TwitterName:            twitterAccountDetails.Name(),
		TwitterUsername:        twitterAccountDetails.Username(),
		TwitterProfileImageURL: twitterAccountDetails.ProfileImageURL(),
	}
}

type transportPublicKey struct {
	Npub string `json:"npub"`
}

func newTransportPublicKey(linkedPublicKey *domain.LinkedPublicKey) transportPublicKey {
	return transportPublicKey{Npub: linkedPublicKey.PublicKey().Npub()}
}

func newTransportPublicKeys(linkedPublicKeys []*domain.LinkedPublicKey) []transportPublicKey {
	result := make([]transportPublicKey, 0) // render empty slice as "[]" not "null"
	for _, v := range linkedPublicKeys {
		result = append(result, newTransportPublicKey(v))
	}
	return result
}
