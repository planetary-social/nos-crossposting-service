package notifications_test

import (
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/notifications"
	"github.com/stretchr/testify/require"
)

func TestGenerator(t *testing.T) {
	testCases := []struct {
		Name string

		EventKind domain.EventKind
	}{
		{
			Name: "note",

			EventKind: domain.EventKindNote,
		},
		{
			Name: "reaction",

			EventKind: domain.EventKindReaction,
		},
		{
			Name: "edm",

			EventKind: domain.EventKindEncryptedDirectMessage,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			logger := logging.NewDevNullLogger()
			g := notifications.NewGenerator(logger)

			pk1, _ := fixtures.SomeKeyPair()
			pk2, sk2 := fixtures.SomeKeyPair()

			libevent := nostr.Event{
				PubKey:    pk2.Hex(),
				CreatedAt: nostr.Timestamp(time.Now().Unix()),
				Kind:      testCase.EventKind.Int(),
				Tags: nostr.Tags{
					nostr.Tag{"p", pk1.Hex()},
				},
				Content: "some content",
			}

			err := libevent.Sign(sk2)
			require.NoError(t, err)

			event, err := domain.NewEvent(libevent)
			require.NoError(t, err)

			token := fixtures.SomeAPNSToken()

			result, err := g.Generate(pk1, token, event)
			require.NoError(t, err)

			require.Len(t, result, 1)

			notification := result[0]
			require.Equal(t,
				`{"aps":{"content-available":1}}`,
				string(notification.Payload()),
			)
			require.Equal(t, token, notification.APNSToken())
			require.Equal(t, event, notification.Event())
		})
	}
}
