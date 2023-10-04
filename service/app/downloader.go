package app

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
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

const (
	getRelaysYoungerThan  = 6 * 30 * 24 * time.Hour
	recheckRelayListEvery = 5 * time.Minute

	reconnectEvery           = 5 * time.Minute
	getPublicKeysYoungerThan = 6 * 30 * 24 * time.Hour
	manageSubscriptionsEvery = 5 * time.Minute

	howFarIntoThePastToLook = 24 * time.Hour

	storeMetricsEvery = 10 * time.Second
)

type ReceivedEventPublisher interface {
	Publish(relay domain.RelayAddress, event domain.Event)
}

type Downloader struct {
	eventWasAlreadySavedCache EventWasAlreadySavedCache
	transactionProvider       TransactionProvider
	receivedEventPublisher    ReceivedEventPublisher
	logger                    logging.Logger
	metrics                   Metrics

	relayDownloaders     map[domain.RelayAddress]*RelayDownloader
	relayDownloadersLock sync.Mutex
}

func NewDownloader(
	eventWasAlreadySavedCache EventWasAlreadySavedCache,
	transaction TransactionProvider,
	receivedEventPublisher ReceivedEventPublisher,
	logger logging.Logger,
	metrics Metrics,
) *Downloader {
	return &Downloader{
		eventWasAlreadySavedCache: eventWasAlreadySavedCache,
		transactionProvider:       transaction,
		receivedEventPublisher:    receivedEventPublisher,
		logger:                    logger.New("downloader"),
		metrics:                   metrics,

		relayDownloaders: map[domain.RelayAddress]*RelayDownloader{},
	}
}

func (d *Downloader) Run(ctx context.Context) error {
	go d.storeMetricsLoop(ctx)

	for {
		if err := d.updateRelays(ctx); err != nil {
			d.logger.Error().
				WithError(err).
				Message("error updating relays")
		}

		select {
		case <-time.After(recheckRelayListEvery):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *Downloader) storeMetricsLoop(ctx context.Context) {
	for {
		d.storeMetrics()

		select {
		case <-time.After(storeMetricsEvery):
		case <-ctx.Done():
			return
		}
	}
}

func (d *Downloader) storeMetrics() {
	d.relayDownloadersLock.Lock()
	defer d.relayDownloadersLock.Unlock()

	v := make(map[RelayDownloaderState]int)

	for _, downloader := range d.relayDownloaders {
		s := downloader.GetState()
		v[s] = v[s] + 1
	}

	for state, n := range v {
		d.metrics.MeasureRelayDownloadersState(n, state)
	}
}

func (d *Downloader) updateRelays(ctx context.Context) error {
	relayAddresses, err := d.getRelays(ctx)
	if err != nil {
		return errors.Wrap(err, "error getting relays")
	}

	d.relayDownloadersLock.Lock()
	defer d.relayDownloadersLock.Unlock()

	for relayAddress, relayDownloader := range d.relayDownloaders {
		if !relayAddresses.Contains(relayAddress) {
			d.logger.Debug().
				WithField("relay", relayAddress.String()).
				Message("deleting a relay downloader")
			delete(d.relayDownloaders, relayAddress)
			relayDownloader.Stop()
		}
	}

	for _, relayAddress := range relayAddresses.List() {
		if _, ok := d.relayDownloaders[relayAddress]; !ok {
			d.logger.Debug().
				WithField("relay", relayAddress.String()).
				Message("creating a relay downloader")
			relayDownloader := NewRelayDownloader(
				ctx,
				d.eventWasAlreadySavedCache,
				d.transactionProvider,
				d.receivedEventPublisher,
				d.logger,
				relayAddress,
			)
			d.relayDownloaders[relayAddress] = relayDownloader
		}
	}

	return nil
}

func (d *Downloader) getRelays(ctx context.Context) (*internal.Set[domain.RelayAddress], error) {
	var relays []domain.RelayAddress

	//if err := d.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
	//	tmp, err := adapters.Relays.GetRelays(ctx, time.Now().Add(-getRelaysYoungerThan))
	//	if err != nil {
	//		return errors.Wrap(err, "error getting relays")
	//	}
	//	relays = tmp
	//	return nil
	//}); err != nil {
	//	return nil, errors.Wrap(err, "transaction error")
	//}

	return internal.NewSet(relays), nil
}

type RelayDownloaderState struct {
	s string
}

func (r RelayDownloaderState) String() string {
	return r.s
}

var (
	RelayDownloaderStateInitializing = RelayDownloaderState{"initializing"}
	RelayDownloaderStateConnected    = RelayDownloaderState{"connected"}
	RelayDownloaderStateDisconnected = RelayDownloaderState{"disconnected"}
)

type RelayDownloader struct {
	eventWasAlreadySavedCache EventWasAlreadySavedCache
	transactionProvider       TransactionProvider
	receivedEventPublisher    ReceivedEventPublisher
	logger                    logging.Logger

	state      RelayDownloaderState
	stateMutex sync.Mutex

	address domain.RelayAddress
	cancel  context.CancelFunc
}

func NewRelayDownloader(
	ctx context.Context,
	eventWasAlreadySavedCache EventWasAlreadySavedCache,
	transactionProvider TransactionProvider,
	receivedEventPublisher ReceivedEventPublisher,
	logger logging.Logger,
	address domain.RelayAddress,
) *RelayDownloader {
	ctx, cancel := context.WithCancel(ctx)
	v := &RelayDownloader{
		eventWasAlreadySavedCache: eventWasAlreadySavedCache,
		transactionProvider:       transactionProvider,
		receivedEventPublisher:    receivedEventPublisher,
		logger:                    logger.New(fmt.Sprintf("relayDownloader(%s)", address)),

		state: RelayDownloaderStateInitializing,

		cancel:  cancel,
		address: address,
	}
	go v.run(ctx)
	return v
}

func (d *RelayDownloader) run(ctx context.Context) {
	for {
		if err := d.connectAndDownload(ctx); err != nil {
			d.logger.Error().
				WithError(err).
				Message("error connecting and downloading")
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(reconnectEvery):
			continue
		}
	}
}

func (d *RelayDownloader) connectAndDownload(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer d.setState(RelayDownloaderStateDisconnected)

	d.logger.Trace().Message("connecting")

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, d.address.String(), nil)
	if err != nil {
		return errors.Wrap(err, "error dialing the relay")
	}

	d.setState(RelayDownloaderStateConnected)

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	go func() {
		if err := d.manageSubs(ctx, conn); err != nil {
			d.logger.Error().
				WithError(err).
				Message("error managing subs")
		}
	}()

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			return errors.Wrap(err, "error reading a message")
		}

		if err := d.handleMessage(messageBytes); err != nil {
			return errors.Wrap(err, "error handling message")
		}
	}
}

func (d *RelayDownloader) handleMessage(messageBytes []byte) error {
	envelope := nostr.ParseMessage(messageBytes)
	if envelope == nil {
		return errors.New("error parsing message, we are never going to find out what error unfortunately due to the design of this library")
	}

	switch v := envelope.(type) {
	case *nostr.EOSEEnvelope:
		d.logger.Trace().
			WithField("subscription", string(*v)).
			Message("received EOSE")
	case *nostr.EventEnvelope:
		event, err := domain.NewEvent(v.Event)
		if err != nil {
			return errors.Wrap(err, "error creating an event")
		}
		if !d.eventWasAlreadySavedCache.EventWasAlreadySaved(event.Id()) {
			d.receivedEventPublisher.Publish(d.address, event)
		}
	default:
		d.logger.
			Debug().
			WithField("message", string(messageBytes)).
			Message("unhandled message")
	}

	return nil
}

func (d *RelayDownloader) setState(state RelayDownloaderState) {
	d.stateMutex.Lock()
	defer d.stateMutex.Unlock()
	d.state = state
}

func (d *RelayDownloader) GetState() RelayDownloaderState {
	d.stateMutex.Lock()
	defer d.stateMutex.Unlock()
	return d.state
}

func (d *RelayDownloader) manageSubs(
	ctx context.Context,
	conn *websocket.Conn,
) error {
	defer conn.Close()

	activeSubscriptions := internal.NewEmptySet[domain.PublicKey]()

	for {
		publicKeys, err := d.getPublicKeys(ctx)
		if err != nil {
			return errors.Wrap(err, "error getting public keys")
		}

		if err := d.updateSubs(conn, activeSubscriptions, publicKeys); err != nil {
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

func (d *RelayDownloader) updateSubs(
	conn *websocket.Conn,
	activeSubscriptions *internal.Set[domain.PublicKey],
	publicKeys *internal.Set[domain.PublicKey],
) error {
	for _, publicKey := range activeSubscriptions.List() {
		if !publicKeys.Contains(publicKey) {
			d.logger.Trace().
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

	for _, publicKey := range publicKeys.List() {
		if ok := activeSubscriptions.Contains(publicKey); !ok {
			d.logger.Trace().
				WithField("publicKey", publicKey).
				Message("opening subscription")

			envelope := d.createRequest(publicKey)

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

func (d *RelayDownloader) createRequest(publicKey domain.PublicKey) nostr.ReqEnvelope {
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

func (d *RelayDownloader) getPublicKeys(ctx context.Context) (*internal.Set[domain.PublicKey], error) {
	var publicKeys []domain.PublicKey

	//if err := d.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
	//	tmp, err := adapters.Relays.GetPublicKeys(ctx, d.address, time.Now().Add(-getPublicKeysYoungerThan))
	//	if err != nil {
	//		return errors.Wrap(err, "error getting public keys")
	//	}
	//	publicKeys = tmp
	//	return nil
	//}); err != nil {
	//	return nil, errors.Wrap(err, "transaction error")
	//}

	return internal.NewSet(publicKeys), nil
}

func (d *RelayDownloader) Stop() {
	d.cancel()
}
