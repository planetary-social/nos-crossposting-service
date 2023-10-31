package domain

import (
	"github.com/boreq/errors"
)

func NoteIsReplyingToOtherEvent(event Event) (bool, error) {
	if event.Kind() != EventKindNote {
		return false, errors.New("incorrect event kind")
	}

	for _, tag := range event.Tags() {
		if tag.IsEvent() {
			return true, nil
		}
	}

	return false, nil
}
