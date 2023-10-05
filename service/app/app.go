package app

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/planetary-social/nos-crossposting-service/service/domain/sessions"
)

var (
	RelayConnectionStateInitializing = RelayConnectionState{"initializing"}
	RelayConnectionStateConnected    = RelayConnectionState{"connected"}
	RelayConnectionStateDisconnected = RelayConnectionState{"disconnected"}
)

var (
	ErrAccountDoesNotExist = errors.New("account doesn't exist")
	ErrSessionDoesNotExist = errors.New("session doesn't exist")
)

type TransactionProvider interface {
	Transact(context.Context, func(context.Context, Adapters) error) error
}

type AccountRepository interface {
	// Returns ErrAccountDoesNotExist.
	GetByTwitterID(twitterID accounts.TwitterID) (*accounts.Account, error)

	// Returns ErrAccountDoesNotExist.
	GetByAccountID(accountID accounts.AccountID) (*accounts.Account, error)

	Save(account *accounts.Account) error
}

type SessionRepository interface {
	// Returns ErrSessionDoesNotExist.
	Get(id sessions.SessionID) (*sessions.Session, error)

	Save(session *sessions.Session) error
}

type PublicKeyRepository interface {
	Save(linkedPublicKey *domain.LinkedPublicKey) error
	List() ([]*domain.LinkedPublicKey, error)
}

type Adapters struct {
	Accounts   AccountRepository
	Sessions   SessionRepository
	PublicKeys PublicKeyRepository
}

type Application struct {
	GetSessionAccount *GetSessionAccountHandler
	LoginOrRegister   *LoginOrRegisterHandler
	LinkPublicKey     *LinkPublicKeyHandler
}

type ReceivedEvent struct {
	relay domain.RelayAddress
	event domain.Event
}

func NewReceivedEvent(relay domain.RelayAddress, event domain.Event) ReceivedEvent {
	return ReceivedEvent{relay: relay, event: event}
}

func (r ReceivedEvent) Relay() domain.RelayAddress {
	return r.relay
}

func (r ReceivedEvent) Event() domain.Event {
	return r.event
}

type ReceivedEventSubscriber interface {
	Subscribe(ctx context.Context) <-chan ReceivedEvent
}

type Metrics interface {
	StartApplicationCall(handlerName string) ApplicationCall
	ReportNumberOfPublicKeyDownloaders(n int)
	ReportNumberOfPublicKeyDownloaderRelays(publicKey domain.PublicKey, n int)
	ReportRelayConnectionState(relayAddress domain.RelayAddress, state RelayConnectionState)
}

type ApplicationCall interface {
	// End accepts a pointer so that you can defer this call without wrapping it
	// in an anonymous function
	End(err *error)
}

type AccountIDGenerator interface {
	GenerateAccountID() (accounts.AccountID, error)
}

type SessionIDGenerator interface {
	GenerateSessionID() (sessions.SessionID, error)
}

type RelayConnectionState struct {
	s string
}

func (r RelayConnectionState) String() string {
	return r.s
}
