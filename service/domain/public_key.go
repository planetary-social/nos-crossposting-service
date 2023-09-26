package domain

import (
	"encoding/hex"
	"github.com/nbd-wtf/go-nostr/nip19"

	"github.com/boreq/errors"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type PublicKey struct {
	s string
}

func NewPublicKeyFromHex(s string) (PublicKey, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return PublicKey{}, errors.Wrap(err, "error decoding hex")
	}

	if len(b) != secp256k1.PrivKeyBytesLen {
		return PublicKey{}, errors.New("invalid length")
	}

	s = hex.EncodeToString(b)
	return PublicKey{s}, nil
}

func NewPublicKeyFromNpub(s string) (PublicKey, error) {
	prefix, hexString, err := nip19.Decode(s)
	if err != nil {
		return PublicKey{}, errors.Wrap(err, "error decoding a nip19 entity")
	}

	if prefix != "npub" {
		return PublicKey{}, errors.New("passed something which isn't an npub")
	}

	return NewPublicKeyFromHex(hexString.(string))
}

func (k PublicKey) Hex() string {
	return k.s
}

func (k PublicKey) Bytes() []byte {
	b, err := hex.DecodeString(k.s)
	if err != nil {
		panic(err)
	}
	return b
}
