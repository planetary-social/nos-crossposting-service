package domain_test

import (
	"testing"

	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/stretchr/testify/require"
)

func TestTweetGenerator_GeneratesTweetsFromNotes(t *testing.T) {
	eventJSON := `{
  "id": "4376c65d2f232afbe9b882a35baa4f6fe8667c4e684749af565f981833ed6a65",
  "pubkey": "6e468422dfb74a5738702a8823b9b28168abab8655faacb6853cd0ee15deee93",
  "created_at": 1673347337,
  "kind": 1,
  "content": "Walled gardens became prisons, and nostr is the first step towards tearing down the prison walls.",
  "tags": [
    ["e", "3da979448d9ba263864c4d6f14984c423a3838364ec255f03c7904b1ae77f206"],
    ["p", "bf2376e17ba4ec269d10fcc996a4746b451152be9031fa48e74553dde5526bce"]
  ],
  "sig": "908a15e46fb4d8675bab026fc230a0e3542bfade63da02d542fb78b2a8513fcd0092619a2c8c1221e581946e0191f2af505dfdf8657a414dbca329186f009262"
}`

	event, err := domain.NewEventFromJSON([]byte(eventJSON))
	require.NoError(t, err)

	g := domain.NewTweetGenerator()
	tweets, err := g.Generate(event)
	require.NoError(t, err)
	require.Equal(t,
		[]domain.Tweet{
			domain.NewTweet(
				"Nostr note made by npub1dergggklka99wwrs92yz8wdjs952h2ux2ha2ed598ngwu9w7a6fsh9xzpc:\n\nWalled gardens became prisons, and nostr is the first step towards tearing down the prison walls.",
			),
		},
		tweets,
	)
}
