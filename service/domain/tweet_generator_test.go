package domain_test

import (
	"fmt"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/stretchr/testify/require"
)

func TestTweetGenerator(t *testing.T) {
	testCases := []struct {
		Name           string
		Event          nostr.Event
		GeneratesTweet bool
	}{
		{
			Name: "not_a_reply",
			Event: nostr.Event{
				Kind: domain.EventKindNote.Int(),
				Tags: []nostr.Tag{
					[]string{"p", fixtures.SomePublicKey().Hex()},
				},
				Content: "Some text.",
			},
			GeneratesTweet: true,
		},
		{
			Name: "reply",
			Event: nostr.Event{
				Kind: domain.EventKindNote.Int(),
				Tags: []nostr.Tag{
					[]string{"p", fixtures.SomePublicKey().Hex()},
					[]string{"e", fixtures.SomeEventID().Hex()},
				},
				Content: "Some text.",
			},
			GeneratesTweet: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			libevent := testCase.Event

			authorPublicKey, authorPrivateKey := fixtures.SomeKeyPair()

			err := libevent.Sign(authorPrivateKey)
			require.NoError(t, err)

			event, err := domain.NewEvent(libevent)
			require.NoError(t, err)

			g := domain.NewTweetGenerator()
			tweets, err := g.Generate(event)
			require.NoError(t, err)

			if testCase.GeneratesTweet {
				require.Equal(t,
					[]domain.Tweet{
						domain.NewTweet(
							fmt.Sprintf(
								"Nostr note made by https://njump.me/%s:\n\n%s",
								authorPublicKey.Npub(),
								event.Content(),
							),
						),
					},
					tweets,
				)
			} else {
				require.Empty(t, tweets)
			}
		})
	}
}
