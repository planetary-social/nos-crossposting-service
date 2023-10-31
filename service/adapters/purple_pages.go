package adapters

import (
	"context"
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
	ErrPurplePagesTimeout             = errors.New("purple pages lookup timed out")
)

var purplePagesAddress = domain.MustNewRelayAddress("wss://purplepag.es")

const purplePagesLookupTimeout = 10 * time.Second

type PurplePages struct {
	logger     logging.Logger
	metrics    app.Metrics
	connection *RelayConnection
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

	var compoundError error
	errorsCounter := 0

	for i := 0; i < 2; i++ {
		result := <-ch

		if err := result.Err; err != nil {
			if errors.Is(err, ErrPurplePagesTimeout) {
				return nil, errors.Wrap(err, "one of the lookups timed out")
			}

			compoundError = multierror.Append(compoundError, err)
			errorsCounter++
			continue
		}
		results.PutMany(result.Addresses)
	}

	if errorsCounter == 2 {
		return nil, compoundError
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
			return nil, ErrRelayListNotFoundInPurplePages
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

	return nil, ErrPurplePagesTimeout
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
			return nil, ErrRelayListNotFoundInPurplePages
		}

		event := eventOrEOSE.Event()

		switch event.Kind() {
		case domain.EventKindContacts:
			result, err := domain.GetRelaysFromContactsEvent(event)
			if err != nil {
				return nil, errors.Wrap(err, "error getting contacts from event")
			}
			return result, nil
		default:
			return nil, errors.New("unexpected event kind")
		}
	}

	return nil, ErrPurplePagesTimeout
}

type relaysOrError struct {
	Err       error
	Addresses []domain.RelayAddress
}
