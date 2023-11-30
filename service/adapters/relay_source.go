package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

const refreshPurplePagesAfter = 30 * time.Minute

var hardcodedRelayAddresses = []domain.RelayAddress{
	domain.MustNewRelayAddress("wss://relay.damus.io"),
	domain.MustNewRelayAddress("wss://nos.lol"),
	domain.MustNewRelayAddress("wss://relay.current.fyi"),
}

type RelaySource struct {
	logger      logging.Logger
	purplePages []*CachedPurplePages
}

func NewRelaySource(logger logging.Logger, purplePages []*CachedPurplePages) *RelaySource {
	return &RelaySource{
		logger:      logger,
		purplePages: purplePages,
	}
}

func (p RelaySource) GetRelays(ctx context.Context, publicKey domain.PublicKey) ([]domain.RelayAddress, error) {
	result := internal.NewSet[domain.RelayAddress](hardcodedRelayAddresses)

	for _, purplePages := range p.purplePages {
		relayAddressesFromPurplePages, err := purplePages.GetRelays(ctx, publicKey)
		if err != nil {
			return nil, fmt.Errorf("error getting relays from '%s'", purplePages.Address().String())
		}
		result.PutMany(relayAddressesFromPurplePages)
	}

	return result.List(), nil
}
