package domain

import (
	"unicode"

	"github.com/boreq/errors"
)

var tagProfile = MustNewEventTagName("p")

func GetMentionsFromTags(tags []EventTag) ([]PublicKey, error) {
	var mentions []PublicKey

	for _, tag := range tags {
		if tag.IsProfile() {
			pubKey, err := tag.Profile()
			if err != nil {
				return nil, errors.Wrapf(err, "error getting public key from tag '%+v'", tag)
			}
			mentions = append(mentions, pubKey)
		}
	}

	return mentions, nil
}

type EventTag struct {
	name EventTagName
	tag  []string
}

func NewEventTag(tag []string) (EventTag, error) {
	if len(tag) < 2 {
		return EventTag{}, errors.New("tag needs at least two fields I recon")
	}

	name, err := NewEventTagName(tag[0])
	if err != nil {
		return EventTag{}, errors.Wrap(err, "invalid tag name")
	}

	return EventTag{name: name, tag: tag}, nil
}

func (e EventTag) Name() EventTagName {
	return e.name
}

func (e EventTag) FirstValue() string {
	return e.tag[1]
}

func (e EventTag) FirstValueIsAnEmptyString() bool {
	return e.FirstValue() == ""
}

func (e EventTag) IsProfile() bool {
	return e.name == tagProfile
}

func (e EventTag) Profile() (PublicKey, error) {
	if !e.IsProfile() {
		return PublicKey{}, errors.New("not a profile tag")
	}
	return NewPublicKeyFromHex(e.tag[1])
}

type EventTagName struct {
	s string
}

func NewEventTagName(s string) (EventTagName, error) {
	if s == "" {
		return EventTagName{}, errors.New("missing tag name")
	}

	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '-' {
			return EventTagName{}, errors.New("tag name should only contain letters and numbers")
		}
	}

	return EventTagName{s}, nil
}

func MustNewEventTagName(s string) EventTagName {
	v, err := NewEventTagName(s)
	if err != nil {
		panic(err)
	}
	return v
}

func (e EventTagName) String() string {
	return e.s
}
