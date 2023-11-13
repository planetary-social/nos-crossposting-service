package domain_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/content"
	"github.com/stretchr/testify/require"
)

func TestTweetGenerator(t *testing.T) {
	testCases := []struct {
		Name            string
		Event           nostr.Event
		ExpectedContent string
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
			ExpectedContent: "Some text.",
		},
		{
			Name: "not_a_reply_long",
			Event: nostr.Event{
				Kind: domain.EventKindNote.Int(),
				Tags: []nostr.Tag{
					[]string{"p", fixtures.SomePublicKey().Hex()},
				},
				Content: strings.Repeat("Some text. ", 100),
			},
			ExpectedContent: "Some text. Some text. Some text. Some text. Some text. Some text. Some text. Some text. Some text. Some text. Some text. Some text. Some text. Some text. Some text. Some text. Some text. Some text....",
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
			ExpectedContent: "",
		},
		{
			Name: "event_with_nostr_link",
			Event: nostr.Event{
				Kind:    domain.EventKindNote.Int(),
				Content: "The content marketing on social can be totally crazy. Just imagine once it is mostly created by LLMs? \n\nnostr:note14aj40jvqs3auq2488c9qxgsqh79zdl0vyhzvzp275g44hhe4etxss9ncxd",
			},
			ExpectedContent: "The content marketing on social can be totally crazy. Just imagine once it is mostly created by LLMs? \n\nhttps://njump.me/note14aj40jvqs3auq2488c9qxgsqh79zdl0vyhzvzp275g44hhe4etxss9ncxd",
		},
		{
			Name: "link_is_not_split",
			Event: nostr.Event{
				Kind:    domain.EventKindNote.Int(),
				Content: strings.Repeat("a", 195) + " https://example.com",
			},
			ExpectedContent: strings.Repeat("a", 195) + " ...",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			libevent := testCase.Event

			_, authorPrivateKey := fixtures.SomeKeyPair()

			err := libevent.Sign(authorPrivateKey)
			require.NoError(t, err)

			event, err := domain.NewEvent(libevent)
			require.NoError(t, err)

			transformer := content.NewTransformer()
			g := domain.NewTweetGenerator(transformer)
			tweets, err := g.Generate(event)
			require.NoError(t, err)

			if testCase.ExpectedContent != "" {
				require.Equal(t,
					[]domain.Tweet{
						domain.NewTweet(
							fmt.Sprintf(
								"%s\n\nhttps://njump.me/%s",
								testCase.ExpectedContent,
								event.Nevent(),
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
