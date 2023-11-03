package domain_test

import (
	"strings"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/stretchr/testify/require"
)

func TestEvent_Nevent(t *testing.T) {
	_, sk := fixtures.SomeKeyPair()

	libevent := nostr.Event{
		Kind:    domain.EventKindNote.Int(),
		Content: "Note text.",
	}
	err := libevent.Sign(sk)
	require.NoError(t, err)

	event, err := domain.NewEvent(libevent)
	require.NoError(t, err)
	require.Greater(t, len(event.Nevent()), 0)
	require.True(t, strings.HasPrefix(event.Nevent(), "nevent"))

	_, v, err := nip19.Decode(event.Nevent())
	require.NoError(t, err)
	readNevent := v.(nostr.EventPointer)
	require.Equal(t, event.Id().Hex(), readNevent.ID)
	require.Equal(t, event.PublicKey().Hex(), readNevent.Author)
}
