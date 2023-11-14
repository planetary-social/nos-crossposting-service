package adapters

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/gorilla/websocket"
	"github.com/nbd-wtf/go-nostr"
	"github.com/oklog/ulid/v2"
	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

type RelayConnection struct {
	address domain.RelayAddress
	logger  logging.Logger

	state      app.RelayConnectionState
	stateMutex sync.Mutex

	subscriptions      map[string]subscription
	subscriptionsMutex sync.Mutex
}

func NewRelayConnection(address domain.RelayAddress, logger logging.Logger) *RelayConnection {
	return &RelayConnection{
		address:       address,
		logger:        logger.New(fmt.Sprintf("relayConnection(%s)", address.String())),
		subscriptions: make(map[string]subscription),
	}
}

func (r *RelayConnection) Run(ctx context.Context) {
	for {
		if err := r.run(ctx); err != nil {
			l := r.logger.Error()
			if r.errorIsCommonAndShouldNotBeLoggedOnErrorLevel(err) {
				l = r.logger.Debug()
			}
			l.WithError(err).Message("encountered an error")
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(reconnectAfter):
			continue
		}
	}
}

func (r *RelayConnection) errorIsCommonAndShouldNotBeLoggedOnErrorLevel(err error) bool {
	if errors.Is(err, DialError{}) {
		return true
	}

	if errors.Is(err, ReadMessageError{}) {
		return true
	}

	return false
}

func (r *RelayConnection) GetEvents(ctx context.Context, publicKey domain.PublicKey, eventKinds []domain.EventKind, maxAge *time.Duration) <-chan app.EventOrEndOfSavedEvents {
	r.subscriptionsMutex.Lock()
	defer r.subscriptionsMutex.Unlock()

	ch := make(chan app.EventOrEndOfSavedEvents)
	uuid := ulid.Make().String()
	r.subscriptions[uuid] = subscription{
		ctx:        ctx,
		ch:         ch,
		uuid:       uuid,
		publicKey:  publicKey,
		eventKinds: eventKinds,
		maxAge:     maxAge,
	}

	go func() {
		<-ctx.Done()
		if err := r.removeChannel(ch); err != nil {
			panic(err)
		}
	}()

	return ch
}

func (r *RelayConnection) State() app.RelayConnectionState {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	return r.state
}

func (r *RelayConnection) Address() domain.RelayAddress {
	return r.address
}

func (r *RelayConnection) removeChannel(chToRemove chan app.EventOrEndOfSavedEvents) error {
	r.subscriptionsMutex.Lock()
	defer r.subscriptionsMutex.Unlock()

	for uuid, subscription := range r.subscriptions {
		if chToRemove == subscription.ch {
			close(subscription.ch)
			delete(r.subscriptions, uuid)
			return nil
		}
	}

	return errors.New("somehow the channel was already removed")
}

func (r *RelayConnection) run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer r.setState(app.RelayConnectionStateDisconnected)

	r.logger.Trace().Message("connecting")

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, r.address.String(), nil)
	if err != nil {
		return NewDialError(err)
	}

	r.setState(app.RelayConnectionStateConnected)
	r.logger.Trace().Message("connected")

	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			r.logger.Debug().WithError(err).Message("error when closing connection due to closed context")
		}
	}()

	go func() {
		if err := r.manageSubs(ctx, conn); err != nil {
			if !errors.Is(err, context.Canceled) {
				r.logger.Error().
					WithError(err).
					Message("error managing subs")
			}
		}
	}()

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			return NewReadMessageError(err)
		}

		if err := r.handleMessage(messageBytes); err != nil {
			return errors.Wrap(err, "error handling message")
		}
	}
}

func (r *RelayConnection) handleMessage(messageBytes []byte) error {
	envelope := nostr.ParseMessage(messageBytes)
	if envelope == nil {
		r.logger.Error().
			WithField("messageBytesAsHex", hex.EncodeToString(messageBytes)).
			Message("error parsing an incoming message")
		return errors.New("error parsing message, we are never going to find out what error unfortunately due to the design of this library")
	}

	switch v := envelope.(type) {
	case *nostr.EOSEEnvelope:
		r.logger.Trace().
			WithField("subscription", string(*v)).
			Message("received EOSE")
		r.passValueToChannel(string(*v), app.NewEventOrEndOfSavedEventsWithEOSE())
	case *nostr.EventEnvelope:
		r.logger.Trace().
			WithField("subscription", *v.SubscriptionID).
			Message("received event")
		event, err := domain.NewEvent(v.Event)
		if err != nil {
			return errors.Wrap(err, "error creating an event")
		}
		r.passValueToChannel(*v.SubscriptionID, app.NewEventOrEndOfSavedEventsWithEvent(event))
	default:
		r.logger.Debug().
			WithField("message", string(messageBytes)).
			Message("unhandled message")
	}

	return nil
}

func (r *RelayConnection) passValueToChannel(uuid string, value app.EventOrEndOfSavedEvents) {
	r.subscriptionsMutex.Lock()
	defer r.subscriptionsMutex.Unlock()

	if sub, ok := r.subscriptions[uuid]; ok {
		select {
		case <-sub.ctx.Done():
		case sub.ch <- value:
		}
	}
}

func (r *RelayConnection) setState(state app.RelayConnectionState) {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	r.state = state
}

func (r *RelayConnection) manageSubs(
	ctx context.Context,
	conn *websocket.Conn,
) error {
	defer conn.Close()

	activeSubscriptions := internal.NewEmptySet[string]()

	for {
		if err := r.updateSubs(conn, activeSubscriptions); err != nil {
			return errors.Wrap(err, "error updating subscriptions")
		}

		select {
		case <-time.After(manageSubscriptionsEvery):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (r *RelayConnection) updateSubs(
	conn *websocket.Conn,
	activeSubscriptions *internal.Set[string],
) error {
	r.subscriptionsMutex.Lock()
	defer r.subscriptionsMutex.Unlock()

	for _, uuid := range activeSubscriptions.List() {
		if _, ok := r.subscriptions[uuid]; !ok {
			r.logger.Trace().
				WithField("uuid", uuid).
				Message("closing subscription")

			envelope := nostr.CloseEnvelope(uuid)

			envelopeJSON, err := envelope.MarshalJSON()
			if err != nil {
				return errors.Wrap(err, "marshaling close envelope failed")
			}

			if err := conn.WriteMessage(websocket.TextMessage, envelopeJSON); err != nil {
				return errors.Wrap(err, "writing close envelope error")
			}

			activeSubscriptions.Delete(uuid)
		}
	}

	for uuid, subscription := range r.subscriptions {
		if ok := activeSubscriptions.Contains(uuid); !ok {
			envelope := r.createRequest(subscription)

			envelopeJSON, err := envelope.MarshalJSON()
			if err != nil {
				return errors.Wrap(err, "marshaling req envelope failed")
			}

			r.logger.Trace().
				WithField("uuid", uuid).
				WithField("payload", string(envelopeJSON)).
				Message("opening subscription")

			if err := conn.WriteMessage(websocket.TextMessage, envelopeJSON); err != nil {
				return errors.Wrap(err, "writing req envelope error")
			}

			activeSubscriptions.Put(uuid)
		}
	}

	return nil
}

func (r *RelayConnection) createRequest(sub subscription) nostr.ReqEnvelope {
	var eventKindsToDownload []int
	for _, eventKind := range sub.eventKinds {
		eventKindsToDownload = append(eventKindsToDownload, eventKind.Int())
	}

	envelope := nostr.ReqEnvelope{
		SubscriptionID: sub.uuid,
		Filters: nostr.Filters{nostr.Filter{
			Authors: []string{
				sub.publicKey.Hex(),
			},
			Kinds: eventKindsToDownload,
		}},
	}

	if sub.maxAge != nil {
		t := nostr.Timestamp(time.Now().Add(-*sub.maxAge).Unix())
		envelope.Filters[0].Since = &t
	}

	return envelope
}

type subscription struct {
	ctx context.Context
	ch  chan app.EventOrEndOfSavedEvents

	uuid       string
	publicKey  domain.PublicKey
	eventKinds []domain.EventKind
	maxAge     *time.Duration
}

type DialError struct {
	underlying error
}

func NewDialError(underlying error) DialError {
	return DialError{underlying: underlying}
}

func (t DialError) Error() string {
	return fmt.Sprintf("error dialing the relay: %s", t.underlying)
}

func (t DialError) Unwrap() error {
	return t.underlying
}

func (t DialError) Is(target error) bool {
	_, ok1 := target.(DialError)
	_, ok2 := target.(*DialError)
	return ok1 || ok2
}

type ReadMessageError struct {
	underlying error
}

func NewReadMessageError(underlying error) ReadMessageError {
	return ReadMessageError{underlying: underlying}
}

func (t ReadMessageError) Error() string {
	return fmt.Sprintf("error reading a message from websocket: %s", t.underlying)
}

func (t ReadMessageError) Unwrap() error {
	return t.underlying
}

func (t ReadMessageError) Is(target error) bool {
	_, ok1 := target.(ReadMessageError)
	_, ok2 := target.(*ReadMessageError)
	return ok1 || ok2
}
