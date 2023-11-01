package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/errors"
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
	purplePages *PurplePages
	cache       *RelayAddressCache
}

func NewRelaySource(logger logging.Logger, purplePages *PurplePages) *RelaySource {
	return &RelaySource{
		logger:      logger,
		purplePages: purplePages,
		cache:       NewRelayAddressCache(),
	}
}

func (p RelaySource) GetRelays(ctx context.Context, publicKey domain.PublicKey) ([]domain.RelayAddress, error) {
	result := internal.NewEmptySet[domain.RelayAddress]()

	for _, relayAddress := range hardcodedRelayAddresses {
		result.Put(relayAddress)
	}

	relayAddressesFromPurplePages, err := p.getRelaysFromPurplePages(ctx, publicKey)
	if err != nil {
		return nil, errors.Wrap(err, "error getting relays from purple pages")
	}

	for _, relayAddress := range relayAddressesFromPurplePages {
		result.Put(relayAddress)
	}

	return result.List(), nil
}

func (p RelaySource) getRelaysFromPurplePages(ctx context.Context, publicKey domain.PublicKey) ([]domain.RelayAddress, error) {
	var previousEntries []domain.RelayAddress

	entry, ok := p.cache.Get(publicKey)
	if ok {
		previousEntries = entry.Addresses
		if time.Since(entry.T) < refreshPurplePagesAfter {
			return previousEntries, nil
		}
	}

	relayAddressesFromPurplePages, err := p.purplePages.GetRelays(ctx, publicKey)
	if err != nil {
		if errors.Is(err, ErrRelayListNotFoundInPurplePages) ||
			errors.Is(err, ErrPurplePagesTimeout) {
			return previousEntries, nil
		}
		return nil, errors.Wrap(err, "error querying purple pages")
	}

	p.cache.Set(publicKey, relayAddressesFromPurplePages)

	return relayAddressesFromPurplePages, nil
}

type RelayAddressCache struct {
	m    map[domain.PublicKey]Entry
	lock sync.Mutex
}

func NewRelayAddressCache() *RelayAddressCache {
	return &RelayAddressCache{m: make(map[domain.PublicKey]Entry)}
}

func (c *RelayAddressCache) Set(publicKey domain.PublicKey, addresses []domain.RelayAddress) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.m[publicKey] = Entry{
		T:         time.Now(),
		Addresses: addresses,
	}
}

func (c *RelayAddressCache) Get(publicKey domain.PublicKey) (Entry, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	v, ok := c.m[publicKey]
	return v, ok
}

type Entry struct {
	T         time.Time
	Addresses []domain.RelayAddress
}
