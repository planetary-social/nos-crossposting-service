package domain

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal"
)

var (
	EventKindMetadata               = MustNewEventKind(0)
	EventKindNote                   = MustNewEventKind(1)
	EventKindReaction               = MustNewEventKind(7)
	EventKindEncryptedDirectMessage = MustNewEventKind(4)
	EventKindRelayListMetadata      = MustNewEventKind(10002)
)

var eventKindsToDownload = internal.NewSet([]EventKind{EventKindNote})

func EventKindsToDownload() []EventKind {
	return eventKindsToDownload.List()
}

func ShouldDownloadEventKind(eventKind EventKind) bool {
	return eventKindsToDownload.Contains(eventKind)
}

type EventKind struct {
	k int
}

func NewEventKind(k int) (EventKind, error) {
	if k < 0 {
		return EventKind{}, errors.New("kind must be positive")
	}
	return EventKind{k}, nil
}

func MustNewEventKind(k int) EventKind {
	v, err := NewEventKind(k)
	if err != nil {
		panic(err)
	}
	return v
}

func (k EventKind) Int() int {
	return k.k
}
