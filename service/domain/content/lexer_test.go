package content_test

import (
	"testing"

	"github.com/planetary-social/nos-crossposting-service/service/domain/content"
	"github.com/stretchr/testify/require"
)

func TestLexer(t *testing.T) {
	testCases := []struct {
		Name string
		In   string
		Out  []content.Token
	}{
		{
			Name: "complex note",
			In:   `This is a nostr note which contains a link: https://example.com. It also has a nostr:nevent1qqs2jy48cze6962ezp9jmj2380v8s9d3ah9ymk7hvvtxsztcruhkjwgzyrnnrjjz0svqt8txvdkalthwk9gp90pdk0xay7u7fjk72pt6d6pw6uue5n2 link and a raw nevent1qqs2jy48cze6962ezp9jmj2380v8s9d3ah9ymk7hvvtxsztcruhkjwgzyrnnrjjz0svqt8txvdkalthwk9gp90pdk0xay7u7fjk72pt6d6pw6uue5n2 link.`,
			Out: []content.Token{
				{
					Type: content.TokenTypeText,
					Text: "This is a nostr note which contains a link: ",
				},
				{
					Type: content.TokenTypeLink,
					Text: "https://example.com",
				},
				{
					Type: content.TokenTypeText,
					Text: ". It also has a ",
				},
				{
					Type: content.TokenTypeNostrLink,
					Text: "nevent1qqs2jy48cze6962ezp9jmj2380v8s9d3ah9ymk7hvvtxsztcruhkjwgzyrnnrjjz0svqt8txvdkalthwk9gp90pdk0xay7u7fjk72pt6d6pw6uue5n2",
				},
				{
					Type: content.TokenTypeText,
					Text: " link and a raw ",
				},
				{
					Type: content.TokenTypeNostrLink,
					Text: "nevent1qqs2jy48cze6962ezp9jmj2380v8s9d3ah9ymk7hvvtxsztcruhkjwgzyrnnrjjz0svqt8txvdkalthwk9gp90pdk0xay7u7fjk72pt6d6pw6uue5n2",
				},
				{
					Type: content.TokenTypeText,
					Text: " link.",
				},
			},
		},
		{
			Name: "malformed_link_with_just_protocol",
			In:   `nostr:`,
			Out: []content.Token{
				{
					Type: content.TokenTypeText,
					Text: "nostr:",
				},
			},
		},
		{
			Name: "malformed_link_with_just_protocol_and_type",
			In:   `nostr:nevent`,
			Out: []content.Token{
				{
					Type: content.TokenTypeText,
					Text: "nostr:nevent",
				},
			},
		},
		{
			Name: "malformed_link_with_incorrect_data",
			In:   `nostr:neventl1Ii`,
			Out: []content.Token{
				{
					Type: content.TokenTypeNostrLink,
					Text: "neventl1",
				},
				{
					Type: content.TokenTypeText,
					Text: "Ii",
				},
			},
		},
		{
			Name: "links_with_protocol",
			In:   `nostr:npubac nostr:neventac nostr:noteac`,
			Out: []content.Token{
				{
					Type: content.TokenTypeNostrLink,
					Text: "npubac",
				},
				{
					Type: content.TokenTypeText,
					Text: " ",
				},
				{
					Type: content.TokenTypeNostrLink,
					Text: "neventac",
				},
				{
					Type: content.TokenTypeText,
					Text: " ",
				},
				{
					Type: content.TokenTypeNostrLink,
					Text: "noteac",
				},
			},
		},
		{
			Name: "links_without_protocol",
			In:   `npubac neventac noteac`,
			Out: []content.Token{
				{
					Type: content.TokenTypeNostrLink,
					Text: "npubac",
				},
				{
					Type: content.TokenTypeText,
					Text: " ",
				},
				{
					Type: content.TokenTypeNostrLink,
					Text: "neventac",
				},
				{
					Type: content.TokenTypeText,
					Text: " ",
				},
				{
					Type: content.TokenTypeNostrLink,
					Text: "noteac",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			l := content.NewLexer(testCase.In)

			tokens, err := l.Lex()
			require.NoError(t, err)

			require.Equal(t, testCase.Out, tokens)
		})
	}
}
