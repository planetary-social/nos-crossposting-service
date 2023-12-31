package domain

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
)

func GetRelaysFromContactsEvent(logger logging.Logger, event Event) ([]RelayAddress, error) {
	if event.Kind() != EventKindContacts {
		return nil, errors.New("incorrect event kind")
	}

	var t map[string]any
	if err := json.Unmarshal([]byte(event.Content()), &t); err != nil {
		return nil, errors.Wrap(err, "json unmarshal error")
	}

	results := internal.NewEmptySet[RelayAddress]()

	for addressString := range t {
		address, err := NewRelayAddress(addressString)
		if err != nil {
			logger.
				Debug().
				WithField("addressString", addressString).
				Message("error creating an address")
			continue
		}
		results.Put(address)
	}

	return results.List(), nil
}
