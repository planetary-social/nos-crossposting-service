package domain

import (
	"strings"

	"github.com/boreq/errors"
)

type RelayAddress struct {
	s string
}

func NewRelayAddress(s string) (RelayAddress, error) {
	if !strings.HasPrefix(s, "ws://") && !strings.HasPrefix(s, "wss://") {
		return RelayAddress{}, errors.New("invalid protocol")
	}

	return RelayAddress{s: s}, nil
}

func MustNewRelayAddress(s string) RelayAddress {
	v, err := NewRelayAddress(s)
	if err != nil {
		panic(err)
	}
	return v
}

func (r RelayAddress) String() string {
	return r.s
}
