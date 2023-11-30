package adapters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

type CachedPurplePages struct {
	logger      logging.Logger
	purplePages *PurplePages
	cache       *RelayAddressCache
}

func NewCachedPurplePages(logger logging.Logger, purplePages *PurplePages) *CachedPurplePages {
	return &CachedPurplePages{
		logger:      logger.New(fmt.Sprintf("CachedPurplePages(%s)", purplePages.Address().String())),
		purplePages: purplePages,
		cache:       NewRelayAddressCache(),
	}
}

func (p CachedPurplePages) GetRelays(ctx context.Context, publicKey domain.PublicKey) ([]domain.RelayAddress, error) {
	entry, ok := p.cache.Get(publicKey)
	if ok {
		if time.Since(entry.T) < refreshPurplePagesAfter {
			return entry.Addresses, nil
		}
	}

	newRelayAddresses, err := p.getRelaysFromPurplePages(ctx, publicKey)
	if err != nil {
		return nil, errors.Wrap(err, "error querying purple pages")
	}

	p.cache.Set(publicKey, newRelayAddresses)
	return newRelayAddresses, nil
}

func (p CachedPurplePages) getRelaysFromPurplePages(ctx context.Context, publicKey domain.PublicKey) ([]domain.RelayAddress, error) {
	relayAddressesFromPurplePages, err := p.purplePages.GetRelays(ctx, publicKey)
	if err != nil {
		if errors.Is(err, ErrRelayListNotFoundInPurplePages) {
			p.logger.Debug().WithError(err).Message("relay list not found in purple pages")
			return nil, nil
		}
		return nil, errors.Wrap(err, "error querying purple pages")
	}
	return relayAddressesFromPurplePages, nil
}

func (p CachedPurplePages) Address() domain.RelayAddress {
	return p.purplePages.Address()
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
