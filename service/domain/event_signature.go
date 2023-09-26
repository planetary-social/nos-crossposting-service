package domain

import (
	"encoding/hex"

	"github.com/boreq/errors"
)

type EventSignature struct {
	s string
}

func NewEventSignature(s string) (EventSignature, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return EventSignature{}, errors.Wrap(err, "error decoding hex")
	}

	if len(b) != 64 {
		return EventSignature{}, errors.New("invalid length")
	}

	s = hex.EncodeToString(b)
	return EventSignature{s}, nil
}

func (sig EventSignature) Hex() string {
	return sig.s
}

func (sig EventSignature) Bytes() []byte {
	b, err := hex.DecodeString(sig.s)
	if err != nil {
		panic(err)
	}
	return b
}
