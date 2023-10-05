package domain

import (
	"encoding/json"
	"strings"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal"
)

// todo make sure that the registration was sent by one of those public keys?
type Registration struct {
	apnsToken APNSToken
	publicKey PublicKey
	relays    []RelayAddress
}

func NewRegistrationFromEvent(event Event) (Registration, error) {
	var v registrationTransport
	if err := json.Unmarshal([]byte(event.Content()), &v); err != nil {
		return Registration{}, errors.Wrap(err, "error unmarshaling content")
	}

	apnsToken, err := NewAPNSTokenFromHex(v.APNSToken)
	if err != nil {
		return Registration{}, errors.Wrap(err, "error creating an apns token")
	}

	publicKey, err := NewPublicKeyFromHex(v.PublicKey)
	if err != nil {
		return Registration{}, errors.Wrap(err, "error creating a public key")
	}

	relays, err := newRelays(v)
	if err != nil {
		return Registration{}, errors.Wrap(err, "error creating relay addresses")
	}

	if event.PubKey() != publicKey {
		return Registration{}, errors.New("public key doesn't match public key from event")
	}

	return Registration{
		apnsToken: apnsToken,
		publicKey: publicKey,
		relays:    relays,
	}, nil
}

func newRelays(v registrationTransport) ([]RelayAddress, error) {
	var relays []RelayAddress
	for _, relayTransport := range v.Relays {
		address, err := NewRelayAddress(relayTransport.Address)
		if err != nil {
			return nil, errors.Wrap(err, "error creating relay address")
		}
		relays = append(relays, address)
	}

	if len(relays) == 0 {
		return nil, errors.New("missing relays")
	}

	return relays, nil
}

func (r Registration) APNSToken() APNSToken {
	return r.apnsToken
}

func (p Registration) PublicKey() PublicKey {
	return p.publicKey
}

func (p Registration) Relays() []RelayAddress {
	return internal.CopySlice(p.relays)
}

type RelayAddress struct {
	s string
}

func NewRelayAddress(s string) (RelayAddress, error) {
	if !strings.HasPrefix(s, "ws://") && !strings.HasPrefix(s, "wss://") {
		return RelayAddress{}, errors.New("invalid protocol")
	}

	// todo validate
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

type registrationTransport struct {
	APNSToken string           `json:"apnsToken"`
	PublicKey string           `json:"publicKey"`
	Relays    []relayTransport `json:"relays"`
}

type relayTransport struct {
	Address string `json:"address"`
}
