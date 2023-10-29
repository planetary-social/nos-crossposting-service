package adapters

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
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
	connection *RelayConnection
}

func NewPurplePages(ctx context.Context, logger logging.Logger) (*PurplePages, error) {
	connection := NewRelayConnection(purplePagesAddress, logger)
	go connection.Run(ctx)

	return &PurplePages{
		logger:     logger,
		connection: connection,
	}, nil
}

func (p *PurplePages) GetRelays(ctx context.Context, publicKey domain.PublicKey) ([]domain.RelayAddress, error) {
	ctx, cancel := context.WithTimeout(ctx, purplePagesLookupTimeout)
	defer cancel()

	for eventOrEOSE := range p.connection.GetEvents(ctx, publicKey, []domain.EventKind{domain.EventKindRelayListMetadata}, nil) {
		if eventOrEOSE.EOSE() {
			return nil, ErrRelayListNotFoundInPurplePages
		}

		if eventOrEOSE.Event().Kind() == domain.EventKindRelayListMetadata {
			var results []domain.RelayAddress
			for _, tag := range eventOrEOSE.Event().Tags() {
				if tag.IsRelay() {
					relayAddress, err := tag.Relay()
					if err != nil {
						return nil, errors.Wrap(err, "error creating a relay address")
					}
					results = append(results, relayAddress)
				}
			}
			return results, nil
		}
	}

	return nil, ErrPurplePagesTimeout
}
