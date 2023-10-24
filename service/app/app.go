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

	Delete(id sessions.SessionID) error
}

type PublicKeyRepository interface {
	Save(linkedPublicKey *domain.LinkedPublicKey) error
	List() ([]*domain.LinkedPublicKey, error)
	ListByPublicKey(publicKey domain.PublicKey) ([]*domain.LinkedPublicKey, error)
	ListByAccountID(accountID accounts.AccountID) ([]*domain.LinkedPublicKey, error)
}

type ProcessedEventRepository interface {
	Save(eventID domain.EventId, twitterID accounts.TwitterID) error
	WasProcessed(eventID domain.EventId, twitterID accounts.TwitterID) (bool, error)
}

type UserTokensRepository interface {
	Save(userTokens *accounts.TwitterUserTokens) error
	Get(id accounts.AccountID) (*accounts.TwitterUserTokens, error)
}

type Publisher interface {
	PublishTweetCreated(accountID accounts.AccountID, tweet domain.Tweet) error
}

type TweetGenerator interface {
	Generate(event domain.Event) ([]domain.Tweet, error)
}

type Twitter interface {
	PostTweet(
		ctx context.Context,
		userAccessToken accounts.TwitterUserAccessToken,
		userAccessSecret accounts.TwitterUserAccessSecret,
		tweet domain.Tweet,
	) error

	GetAccountDetails(
		ctx context.Context,
		userAccessToken accounts.TwitterUserAccessToken,
		userAccessSecret accounts.TwitterUserAccessSecret,
	) (TwitterAccountDetails, error)
}

type TwitterAccountDetailsCache interface {
	Get(accountID accounts.AccountID, updateFn func() (TwitterAccountDetails, error)) (TwitterAccountDetails, error)
}

type Adapters struct {
	Accounts        AccountRepository
	Sessions        SessionRepository
	PublicKeys      PublicKeyRepository
	ProcessedEvents ProcessedEventRepository
	UserTokens      UserTokensRepository
	Publisher       Publisher
}

type Application struct {
	GetSessionAccount        *GetSessionAccountHandler
	GetAccountPublicKeys     *GetAccountPublicKeysHandler
	GetTwitterAccountDetails *GetTwitterAccountDetailsHandler

	LoginOrRegister *LoginOrRegisterHandler
	Logout          *LogoutHandler
	LinkPublicKey   *LinkPublicKeyHandler
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
	ReportCallingTwitterAPIToPostATweet(err error)
	ReportCallingTwitterAPIToGetAUser(err error)
	ReportSubscriptionQueueLength(topic string, n int)
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

type TwitterAccountDetails struct {
	name            string
	username        string
	profileImageURL string
}

func NewTwitterAccountDetails(name string, username string, profileImageURL string) (TwitterAccountDetails, error) {
	if name == "" {
		return TwitterAccountDetails{}, errors.New("name can't be empty")
	}
	if username == "" {
		return TwitterAccountDetails{}, errors.New("username can't be empty")
	}
	return TwitterAccountDetails{
		name:            name,
		username:        username,
		profileImageURL: profileImageURL,
	}, nil
}

func (t TwitterAccountDetails) Name() string {
	return t.name
}

func (t TwitterAccountDetails) Username() string {
	return t.username
}

func (t TwitterAccountDetails) ProfileImageURL() string {
	return t.profileImageURL
}
