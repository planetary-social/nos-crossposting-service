package adapters

import (
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

const cacheTwitterAccountDetailsFor = 15 * time.Minute

type TwitterAccountDetailsCache struct {
	entries     map[string]twitterAccountDetailsCacheEntry // todo cleanup periodically
	entriesLock sync.Mutex
}

func NewTwitterAccountDetailsCache() *TwitterAccountDetailsCache {
	return &TwitterAccountDetailsCache{
		entries: make(map[string]twitterAccountDetailsCacheEntry),
	}
}

func (t *TwitterAccountDetailsCache) Get(accountID accounts.AccountID, updateFn func() (app.TwitterAccountDetails, error)) (app.TwitterAccountDetails, error) {
	t.entriesLock.Lock()
	defer t.entriesLock.Unlock()

	entry, ok := t.entries[accountID.String()]
	if ok && entry.isUpToDate() {
		return entry.value, nil
	}

	newValue, err := updateFn()
	if err != nil {
		return app.TwitterAccountDetails{}, errors.Wrap(err, "error getting new twitter account details")
	}

	t.entries[accountID.String()] = newTwitterAccountDetailsCacheEntry(newValue)
	return newValue, nil
}

type twitterAccountDetailsCacheEntry struct {
	t     time.Time
	value app.TwitterAccountDetails
}

func newTwitterAccountDetailsCacheEntry(value app.TwitterAccountDetails) twitterAccountDetailsCacheEntry {
	return twitterAccountDetailsCacheEntry{t: time.Now(), value: value}
}

func (e *twitterAccountDetailsCacheEntry) isUpToDate() bool {
	return time.Since(e.t) < cacheTwitterAccountDetailsFor
}
