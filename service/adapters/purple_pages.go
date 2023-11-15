package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

var (
	ErrRelayListNotFoundInPurplePages = errors.New("relay list not found in purple pages")

	errLookupFoundNoEvents = errors.New("lookup found no events")
)

var purplePagesAddress = domain.MustNewRelayAddress("wss://purplepag.es")

const purplePagesLookupTimeout = 10 * time.Second

const numLookups = 2

type PurplePages struct {
	logger     logging.Logger
	metrics    app.Metrics
	connection *RelayConnection
	mutex      sync.Mutex // purple pages isn't happy when we open too many concurrent requests
}

func NewPurplePages(
	ctx context.Context,
	logger logging.Logger,
	metrics app.Metrics,
) (*PurplePages, error) {
	connection := NewRelayConnection(purplePagesAddress, logger)
	go connection.Run(ctx)

	return &PurplePages{
		logger:     logger,
		metrics:    metrics,
		connection: connection,
	}, nil
}

func (p *PurplePages) GetRelays(ctx context.Context, publicKey domain.PublicKey) (result []domain.RelayAddress, err error) {
	defer p.metrics.ReportPurplePagesLookupResult(&err)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan relaysOrError)

	go func() {
		addresses, err := p.getRelaysFromContacts(ctx, publicKey)
		select {
		case ch <- relaysOrError{
			Err:       err,
			Addresses: addresses,
		}:
		case <-ctx.Done():
		}
	}()

	go func() {
		addresses, err := p.getRelaysFromRelayMetadata(ctx, publicKey)
		select {
		case ch <- relaysOrError{
			Err:       err,
			Addresses: addresses,
		}:
		case <-ctx.Done():
		}
	}()

	results := internal.NewEmptySet[domain.RelayAddress]()

	var compoundError *multierror.Error

	for i := 0; i < numLookups; i++ {
		result := <-ch
		if err := result.Err; err != nil {
			if !errors.Is(err, errLookupFoundNoEvents) {
				return nil, errors.Wrap(err, "one of the lookups failed")
			}
			compoundError = multierror.Append(compoundError, err)
			continue
		}
		results.PutMany(result.Addresses)
	}

	// if any of the lookups failed with err != errLookupFoundNoEvents return compoundErr
	// if all of the lookups failed with err == errLookupFoundNoEvents return ErrRelayListNotFoundInPurplePages
	if compoundError != nil {
		errLookupFoundNoEventsCounter := 0
		for _, componentErr := range compoundError.Errors {
			if !errors.Is(componentErr, errLookupFoundNoEvents) {
				return nil, errors.Wrap(compoundError, "some lookups failed")
			}
			errLookupFoundNoEventsCounter++
		}

		if errLookupFoundNoEventsCounter == numLookups {
			return nil, ErrRelayListNotFoundInPurplePages
		}
	}

	return results.List(), nil
}

func (p *PurplePages) getRelaysFromRelayMetadata(ctx context.Context, publicKey domain.PublicKey) ([]domain.RelayAddress, error) {
	ctx, cancel := context.WithTimeout(ctx, purplePagesLookupTimeout)
	defer cancel()

	for eventOrEOSE := range p.connection.GetEvents(
		ctx,
		publicKey,
		[]domain.EventKind{
			domain.EventKindRelayListMetadata,
		},
		nil,
	) {
		if eventOrEOSE.EOSE() {
			return nil, errLookupFoundNoEvents
		}

		event := eventOrEOSE.Event()

		switch event.Kind() {
		case domain.EventKindRelayListMetadata:
			var results []domain.RelayAddress
			for _, tag := range event.Tags() {
				if tag.IsRelay() {
					relayAddress, err := tag.Relay()
					if err != nil {
						return nil, errors.Wrap(err, "error creating a relay address")
					}
					results = append(results, relayAddress)
				}
			}
			return results, nil
		default:
			return nil, errors.New("unexpected event kind")
		}
	}

	return nil, errors.New("timeout")
}

func (p *PurplePages) getRelaysFromContacts(ctx context.Context, publicKey domain.PublicKey) ([]domain.RelayAddress, error) {
	ctx, cancel := context.WithTimeout(ctx, purplePagesLookupTimeout)
	defer cancel()

	for eventOrEOSE := range p.connection.GetEvents(
		ctx,
		publicKey,
		[]domain.EventKind{
			domain.EventKindContacts,
		},
		nil,
	) {
		if eventOrEOSE.EOSE() {
			return nil, errLookupFoundNoEvents
		}

		event := eventOrEOSE.Event()

		switch event.Kind() {
		case domain.EventKindContacts:
			result, err := domain.GetRelaysFromContactsEvent(event)
			if err != nil {
				p.logger.
					Debug().
					WithError(err).
					WithField("eventContent", event.Content()).
					Message("error extracting relays from contacts event")
				return nil, errLookupFoundNoEvents
			}
			return result, nil
		default:
			return nil, errors.New("unexpected event kind")
		}
	}

	return nil, errors.New("timeout")
}

type relaysOrError struct {
	Err       error
	Addresses []domain.RelayAddress
}
