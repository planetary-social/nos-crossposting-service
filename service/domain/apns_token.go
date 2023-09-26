package domain

import (
	"encoding/hex"

	"github.com/boreq/errors"
)

type APNSToken struct {
	s string
}

func NewAPNSTokenFromHex(s string) (APNSToken, error) {
	if s == "" {
		return APNSToken{}, errors.New("apns token can't be empty")
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return APNSToken{}, errors.Wrap(err, "error decoding hex")
	}

	s = hex.EncodeToString(b)
	return APNSToken{s}, nil
}

func (t APNSToken) Hex() string {
	return t.s
}
