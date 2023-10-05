package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

const (
	howFarIntoThePastToLook               = 24 * time.Hour
	storeMetricsEvery                     = 30 * time.Second
	refreshDownloaderPublicKeysEvery      = 5 * time.Minute
	refreshPublicKeyDownloaderRelaysEvery = 60 * time.Minute
)

type ReceivedEventPublisher interface {
	Publish(relay domain.RelayAddress, event domain.Event)
}

type RelaySource interface {
	GetRelays(ctx context.Context, publicKey domain.PublicKey) ([]domain.RelayAddress, error)
}

type RelayEventDownloader interface {
	GetEvents(ctx context.Context, publicKey domain.PublicKey, relayAddress domain.RelayAddress, eventKinds []domain.EventKind, maxAge *time.Duration) <-chan EventOrEndOfSavedEvents
}

type Downloader struct {
	transactionProvider    TransactionProvider
	receivedEventPublisher ReceivedEventPublisher
	logger                 logging.Logger
	metrics                Metrics
	relaySource            RelaySource
	relayEventDownloader   RelayEventDownloader

	publicKeyDownloaders     map[domain.PublicKey]context.CancelFunc
	publicKeyDownloadersLock sync.Mutex
}

func NewDownloader(
	transaction TransactionProvider,
	receivedEventPublisher ReceivedEventPublisher,
	logger logging.Logger,
	metrics Metrics,
	relaySource RelaySource,
	relayEventDownloader RelayEventDownloader,
) *Downloader {
	return &Downloader{
		transactionProvider:    transaction,
		receivedEventPublisher: receivedEventPublisher,
		logger:                 logger.New("downloader"),
		metrics:                metrics,
		relaySource:            relaySource,
		relayEventDownloader:   relayEventDownloader,

		publicKeyDownloaders: make(map[domain.PublicKey]context.CancelFunc),
	}
}

func (d *Downloader) Run(ctx context.Context) error {
	go d.storeMetricsLoop(ctx)

	for {
		if err := d.updateDownloaders(ctx); err != nil {
			d.logger.Error().
				WithError(err).
				Message("error updating relays")
		}

		select {
		case <-time.After(refreshDownloaderPublicKeysEvery):
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
	d.publicKeyDownloadersLock.Lock()
	defer d.publicKeyDownloadersLock.Unlock()

	d.metrics.ReportNumberOfPublicKeyDownloaders(len(d.publicKeyDownloaders))
}

func (d *Downloader) updateDownloaders(ctx context.Context) error {
	publicKeys, err := d.getPublicKeys(ctx)
	if err != nil {
		return errors.Wrap(err, "error getting public keys")
	}

	d.publicKeyDownloadersLock.Lock()
	defer d.publicKeyDownloadersLock.Unlock()

	for publicKey, cancelFn := range d.publicKeyDownloaders {
		if !publicKeys.Contains(publicKey) {
			d.logger.Debug().
				WithField("publicKey", publicKey.Hex()).
				Message("stopping a downloader")
			delete(d.publicKeyDownloaders, publicKey)
			cancelFn()
		}
	}

	for _, publicKey := range publicKeys.List() {
		if _, ok := d.publicKeyDownloaders[publicKey]; !ok {
			d.logger.Debug().
				WithField("publicKey", publicKey.Hex()).
				Message("creating a downloader")

			downloader := NewPublicKeyDownloader(
				d.receivedEventPublisher,
				d.relaySource,
				d.relayEventDownloader,
				d.metrics,
				d.logger,
				publicKey,
			)

			ctx, cancel := context.WithCancel(ctx)
			go downloader.Run(ctx)
			d.publicKeyDownloaders[publicKey] = cancel
		}
	}

	return nil
}

func (d *Downloader) getPublicKeys(ctx context.Context) (*internal.Set[domain.PublicKey], error) {
	result := internal.NewEmptySet[domain.PublicKey]()

	if err := d.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		linkedPublicKeys, err := adapters.PublicKeys.List()
		if err != nil {
			return errors.Wrap(err, "error listing public keys")
		}
		for _, linkedPublicKey := range linkedPublicKeys {
			result.Put(linkedPublicKey.PublicKey())
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction error")
	}

	return result, nil
}

type PublicKeyDownloader struct {
	receivedEventPublisher ReceivedEventPublisher
	relaySource            RelaySource
	relayEventDownloader   RelayEventDownloader
	metrics                Metrics
	logger                 logging.Logger

	publicKey domain.PublicKey

	downloaders     map[domain.RelayAddress]context.CancelFunc
	downloadersLock sync.Mutex
}

func NewPublicKeyDownloader(
	receivedEventPublisher ReceivedEventPublisher,
	relaySource RelaySource,
	relayEventDownloader RelayEventDownloader,
	metrics Metrics,
	logger logging.Logger,
	publicKey domain.PublicKey,
) *PublicKeyDownloader {
	v := &PublicKeyDownloader{
		receivedEventPublisher: receivedEventPublisher,
		relaySource:            relaySource,
		relayEventDownloader:   relayEventDownloader,
		metrics:                metrics,
		logger:                 logger.New(fmt.Sprintf("publicKeyDownloader(%s)", publicKey)),

		publicKey: publicKey,
	}
	return v
}

func (d *PublicKeyDownloader) Run(ctx context.Context) {
	go d.storeMetricsLoop(ctx)

	for {
		if err := d.refreshRelays(ctx); err != nil {
			d.logger.Error().
				WithError(err).
				Message("error connecting and downloading")
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(refreshPublicKeyDownloaderRelaysEvery):
			continue
		}
	}
}

func (d *PublicKeyDownloader) storeMetricsLoop(ctx context.Context) {
	for {
		d.storeMetrics()

		select {
		case <-time.After(storeMetricsEvery):
		case <-ctx.Done():
			return
		}
	}
}

func (d *PublicKeyDownloader) storeMetrics() {
	d.downloadersLock.Lock()
	defer d.downloadersLock.Unlock()

	d.metrics.ReportNumberOfPublicKeyDownloaderRelays(d.publicKey, len(d.downloaders))
}

func (d *PublicKeyDownloader) refreshRelays(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	relayAddressess, err := d.relaySource.GetRelays(ctx, d.publicKey)
	if err != nil {
		return errors.Wrap(err, "error getting relayAddressess")
	}

	relayAddressesSet := internal.NewSet(relayAddressess)

	d.downloadersLock.Lock()
	defer d.downloadersLock.Unlock()

	for relayAddress, cancelFn := range d.downloaders {
		if !relayAddressesSet.Contains(relayAddress) {
			d.logger.Debug().
				WithField("relayAddress", relayAddress.String()).
				Message("stopping a downloader")
			delete(d.downloaders, relayAddress)
			cancelFn()
		}
	}

	for _, relayAddress := range relayAddressesSet.List() {
		if _, ok := d.downloaders[relayAddress]; !ok {
			d.logger.Debug().
				WithField("relayAddress", relayAddress.String()).
				Message("creating a downloader")

			ctx, cancel := context.WithCancel(ctx)
			go d.downloadMessages(ctx, relayAddress)
			d.downloaders[relayAddress] = cancel
		}
	}

	return nil
}

func (d *PublicKeyDownloader) downloadMessages(ctx context.Context, relayAddress domain.RelayAddress) {
	t := howFarIntoThePastToLook
	for eventOrEOSE := range d.relayEventDownloader.GetEvents(ctx, d.publicKey, relayAddress, domain.EventKindsToDownload(), &t) {
		if !eventOrEOSE.EOSE() {
			d.receivedEventPublisher.Publish(relayAddress, eventOrEOSE.Event())
		}
	}
}

type EventOrEndOfSavedEvents struct {
	event domain.Event
	eose  bool
}

func NewEventOrEndOfSavedEventsWithEvent(event domain.Event) EventOrEndOfSavedEvents {
	return EventOrEndOfSavedEvents{event: event}
}

func NewEventOrEndOfSavedEventsWithEOSE() EventOrEndOfSavedEvents {
	return EventOrEndOfSavedEvents{eose: true}
}

func (e *EventOrEndOfSavedEvents) Event() domain.Event {
	return e.event
}

func (e *EventOrEndOfSavedEvents) EOSE() bool {
	return e.eose
}
