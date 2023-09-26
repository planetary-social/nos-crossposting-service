package domain

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/boreq/errors"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type EventId struct {
	s string
}

func NewEventId(s string) (EventId, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return EventId{}, errors.Wrap(err, "error decoding hex")
	}

	if len(b) != sha256.Size {
		return EventId{}, errors.New("invalid length")
	}

	s = hex.EncodeToString(b)
	return EventId{s}, nil
}

func MustNewEventId(s string) EventId {
	v, err := NewEventId(s)
	if err != nil {
		panic(err)
	}
	return v
}

func NewEventIdFromNote(s string) (EventId, error) {
	prefix, value, err := nip19.Decode(s)
	if err != nil {
		return EventId{}, errors.Wrap(err, "error calling nip19 decode")
	}
	if prefix != "note" {
		return EventId{}, errors.New("invalid prefix")
	}
	s, ok := value.(string)
	if !ok {
		return EventId{}, errors.New("library returned invalid type")
	}
	return NewEventId(s)
}

func (id EventId) Hex() string {
	return id.s
}

func (id EventId) Bytes() []byte {
	b, err := hex.DecodeString(id.s)
	if err != nil {
		panic(err)
	}
	return b
}
