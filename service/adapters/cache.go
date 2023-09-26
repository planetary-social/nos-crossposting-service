package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

const (
	cacheSavedEventsFor = 60 * time.Minute
	cleanupEvery        = 5 * time.Minute
)

type MemoryEventWasAlreadySavedCache struct {
	cacheLock sync.Mutex
	cache     map[domain.EventId]time.Time
}

func NewMemoryEventWasAlreadySavedCache() *MemoryEventWasAlreadySavedCache {
	return &MemoryEventWasAlreadySavedCache{
		cache: make(map[domain.EventId]time.Time),
	}
}

func (m *MemoryEventWasAlreadySavedCache) Run(ctx context.Context) error {
	for {
		m.cleanup()

		select {
		case <-time.After(cleanupEvery):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *MemoryEventWasAlreadySavedCache) cleanup() {
	m.cacheLock.Lock()
	defer m.cacheLock.Unlock()

	for id, timestamp := range m.cache {
		if time.Since(timestamp) > cacheSavedEventsFor {
			delete(m.cache, id)
		}
	}
}

func (m *MemoryEventWasAlreadySavedCache) MarkEventAsAlreadySaved(id domain.EventId) {
	m.cacheLock.Lock()
	defer m.cacheLock.Unlock()

	m.cache[id] = time.Now()
}

func (m *MemoryEventWasAlreadySavedCache) EventWasAlreadySaved(id domain.EventId) bool {
	m.cacheLock.Lock()
	defer m.cacheLock.Unlock()

	_, ok := m.cache[id]
	return ok
}
