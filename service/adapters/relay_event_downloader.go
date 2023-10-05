package adapters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/gorilla/websocket"
	"github.com/nbd-wtf/go-nostr"
	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

const (
	reconnectAfter           = 1 * time.Minute
	howFarIntoThePastToLook  = 24 * time.Hour
	manageSubscriptionsEvery = 10 * time.Second
)

type RelayEventDownloader struct {
	logger logging.Logger

	ctx context.Context

	connections     map[domain.RelayAddress]*RelayConnection
	connectionsLock sync.Mutex
}

func NewRelayEventDownloader(ctx context.Context, logger logging.Logger) *RelayEventDownloader {
	return &RelayEventDownloader{
		logger:          logger.New("relayEventDownloader"),
		ctx:             ctx,
		connections:     make(map[domain.RelayAddress]*RelayConnection),
		connectionsLock: sync.Mutex{},
	}
}

func (r *RelayEventDownloader) GetEvents(ctx context.Context, publicKey domain.PublicKey, relayAddress domain.RelayAddress) <-chan domain.Event {
	connection := r.getConnection(relayAddress)
	return connection.GetEvents(ctx, publicKey)
}

func (r *RelayEventDownloader) getConnection(relayAddress domain.RelayAddress) *RelayConnection {
	r.connectionsLock.Lock()
	defer r.connectionsLock.Unlock()

	if connection, ok := r.connections[relayAddress]; ok {
		return connection
	}

	connection := NewRelayConnection(relayAddress, r.logger)
	go connection.Run(r.ctx)

	r.connections[relayAddress] = connection
	return connection
}

type RelayConnection struct {
	address domain.RelayAddress
	logger  logging.Logger

	state      app.RelayConnectionState
	stateMutex sync.Mutex

	channels      map[domain.PublicKey][]chWithCtx
	channelsMutex sync.Mutex
}

func NewRelayConnection(address domain.RelayAddress, logger logging.Logger) *RelayConnection {
	return &RelayConnection{
		address: address,
		logger:  logger.New(fmt.Sprintf("relayConnection(%s)", address.String())),
	}
}

func (r *RelayConnection) GetEvents(ctx context.Context, publicKey domain.PublicKey) <-chan domain.Event {
	r.channelsMutex.Lock()
	defer r.channelsMutex.Unlock()

	ch := make(chan domain.Event)
	r.channels[publicKey] = append(r.channels[publicKey], chWithCtx{
		ctx: ctx,
		ch:  ch,
	})

	go func() {
		<-ctx.Done()
		if err := r.removeChannel(publicKey, ch); err != nil {
			panic(err)
		}
	}()

	return ch
}

func (r *RelayConnection) removeChannel(publicKey domain.PublicKey, chToRemove chan domain.Event) error {
	r.channelsMutex.Lock()
	defer r.channelsMutex.Unlock()

	for i, chWithCtx := range r.channels[publicKey] {
		if chToRemove == chWithCtx.ch {
			r.channels[publicKey] = append(r.channels[publicKey][:i], r.channels[publicKey][i+1:]...)
			return nil
		}
	}

	if len(r.channels[publicKey]) == 0 {
		delete(r.channels, publicKey)
	}

	return errors.New("somehow the channel was already removed")
}

func (r *RelayConnection) Run(ctx context.Context) {
	for {
		if err := r.run(ctx); err != nil {
			r.logger.Error().WithError(err).Message("encountered an error")
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(reconnectAfter):
			continue
		}
	}
}

func (r *RelayConnection) run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer r.setState(app.RelayConnectionStateDisconnected)

	r.logger.Trace().Message("connecting")

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, r.address.String(), nil)
	if err != nil {
		return errors.Wrap(err, "error dialing the relay")
	}

	r.setState(app.RelayConnectionStateConnected)

	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			r.logger.Debug().WithError(err).Message("error when closing connection due to closed context")
		}
	}()

	go func() {
		if err := r.manageSubs(ctx, conn); err != nil {
			r.logger.Error().
				WithError(err).
				Message("error managing subs")
		}
	}()

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			return errors.Wrap(err, "error reading a message")
		}

		if err := r.handleMessage(messageBytes); err != nil {
			return errors.Wrap(err, "error handling message")
		}
	}
}

func (r *RelayConnection) handleMessage(messageBytes []byte) error {
	envelope := nostr.ParseMessage(messageBytes)
	if envelope == nil {
		return errors.New("error parsing message, we are never going to find out what error unfortunately due to the design of this library")
	}

	switch v := envelope.(type) {
	case *nostr.EOSEEnvelope:
		r.logger.Trace().
			WithField("subscription", string(*v)).
			Message("received EOSE")
	case *nostr.EventEnvelope:
		event, err := domain.NewEvent(v.Event)
		if err != nil {
			return errors.Wrap(err, "error creating an event")
		}
		r.passEventToAllChannels(event)
	default:
		r.logger.Debug().
			WithField("message", string(messageBytes)).
			Message("unhandled message")
	}

	return nil
}

func (r *RelayConnection) passEventToAllChannels(event domain.Event) {
	r.channelsMutex.Lock()
	defer r.channelsMutex.Unlock()

	for _, chWithCtx := range r.channels[event.PubKey()] {
		select {
		case <-chWithCtx.ctx.Done():
			continue
		case chWithCtx.ch <- event:
			continue
		}
	}
}

func (r *RelayConnection) setState(state app.RelayConnectionState) {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	r.state = state
}

func (r *RelayConnection) GetState() app.RelayConnectionState {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	return r.state
}

func (r *RelayConnection) manageSubs(
	ctx context.Context,
	conn *websocket.Conn,
) error {
	defer conn.Close()

	activeSubscriptions := internal.NewEmptySet[domain.PublicKey]()

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
	activeSubscriptions *internal.Set[domain.PublicKey],
) error {
	r.channelsMutex.Lock()
	defer r.channelsMutex.Unlock()

	for _, publicKey := range activeSubscriptions.List() {
		if _, ok := r.channels[publicKey]; !ok {
			r.logger.Trace().
				WithField("publicKey", publicKey).
				Message("closing subscription")

			envelope := nostr.CloseEnvelope(publicKey.Hex())

			envelopeJSON, err := envelope.MarshalJSON()
			if err != nil {
				return errors.Wrap(err, "marshaling close envelope failed")
			}

			if err := conn.WriteMessage(websocket.TextMessage, envelopeJSON); err != nil {
				return errors.Wrap(err, "writing close envelope error")
			}

			activeSubscriptions.Delete(publicKey)
		}
	}

	for publicKey := range r.channels {
		if ok := activeSubscriptions.Contains(publicKey); !ok {
			r.logger.Trace().
				WithField("publicKey", publicKey).
				Message("opening subscription")

			envelope := r.createRequest(publicKey)

			envelopeJSON, err := envelope.MarshalJSON()
			if err != nil {
				return errors.Wrap(err, "marshaling req envelope failed")
			}

			if err := conn.WriteMessage(websocket.TextMessage, envelopeJSON); err != nil {
				return errors.Wrap(err, "writing req envelope error")
			}

			activeSubscriptions.Put(publicKey)
		}
	}

	return nil
}

func (r *RelayConnection) createRequest(publicKey domain.PublicKey) nostr.ReqEnvelope {
	t := nostr.Timestamp(time.Now().Add(-howFarIntoThePastToLook).Unix())

	var eventKindsToDownload []int
	for _, eventKind := range domain.EventKindsToDownload() {
		eventKindsToDownload = append(eventKindsToDownload, eventKind.Int())
	}

	envelope := nostr.ReqEnvelope{
		SubscriptionID: publicKey.Hex(),
		Filters: nostr.Filters{nostr.Filter{
			Kinds: eventKindsToDownload,
			Tags: map[string][]string{
				"p": {publicKey.Hex()},
			},
			Since: &t,
		}},
	}

	return envelope
}

type chWithCtx struct {
	ctx context.Context
	ch  chan domain.Event
}
