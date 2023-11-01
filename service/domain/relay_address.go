package domain

import (
	"strings"

	"github.com/boreq/errors"
)

const (
	protocolWs  = "ws://"
	protocolWss = "wss://"
)

type RelayAddress struct {
	s string
}

func NewRelayAddress(s string) (RelayAddress, error) {
	if !strings.HasPrefix(s, protocolWs) && !strings.HasPrefix(s, protocolWss) {
		return RelayAddress{}, errors.New("invalid protocol")
	}

	if s == protocolWs || s == protocolWss {
		return RelayAddress{}, errors.New("just protocol")
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

func NormalizeRelayAddress(relayAddress RelayAddress) (RelayAddress, error) {
	addr := strings.TrimSuffix(relayAddress.String(), "/")
	return NewRelayAddress(addr)
}
