package adapters

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

var hardcodedRelayAddresses = []domain.RelayAddress{
	domain.MustNewRelayAddress("wss://relay.damus.io"),
	domain.MustNewRelayAddress("wss://nos.lol"),
	domain.MustNewRelayAddress("wss://relay.current.fyi"),
}

type RelaySource struct {
	logger      logging.Logger
	purplePages *PurplePages
}

func NewRelaySource(logger logging.Logger, purplePages *PurplePages) *RelaySource {
	return &RelaySource{logger: logger, purplePages: purplePages}
}

func (p RelaySource) GetRelays(ctx context.Context, publicKey domain.PublicKey) ([]domain.RelayAddress, error) {
	result := internal.NewEmptySet[domain.RelayAddress]()

	for _, relayAddress := range hardcodedRelayAddresses {
		result.Put(relayAddress)
	}

	relayAddressesFromPurplePages, err := p.purplePages.GetRelays(ctx, publicKey)
	if err != nil {
		if errors.Is(err, ErrRelayListNotFoundInPurplePages) {
			return result.List(), nil
		}
		return nil, errors.Wrap(err, "error querying purple pages")
	}

	for _, relayAddress := range relayAddressesFromPurplePages {
		result.Put(relayAddress)
	}

	return result.List(), nil
}
