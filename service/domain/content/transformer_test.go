package content_test

import (
	"testing"

	"github.com/planetary-social/nos-crossposting-service/service/domain/content"
	"github.com/stretchr/testify/require"
)

func TestTransformer(t *testing.T) {
	testCases := []struct {
		Name string
		In   string
		Out  []content.Element
	}{
		{
			Name: "complex_note",
			In:   `This is a nostr note which contains a link: https://example.com. It also has a nostr:nevent1qqs2jy48cze6962ezp9jmj2380v8s9d3ah9ymk7hvvtxsztcruhkjwgzyrnnrjjz0svqt8txvdkalthwk9gp90pdk0xay7u7fjk72pt6d6pw6uue5n2 link.`,
			Out: []content.Element{
				{
					Type: content.ElementTypeText,
					Text: "This is a nostr note which contains a link: ",
				},
				{
					Type: content.ElementTypeLink,
					Text: "https://example.com",
				},
				{
					Type: content.ElementTypeText,
					Text: ". It also has a ",
				},
				{
					Type: content.ElementTypeLink,
					Text: "https://njump.me/nevent1qqs2jy48cze6962ezp9jmj2380v8s9d3ah9ymk7hvvtxsztcruhkjwgzyrnnrjjz0svqt8txvdkalthwk9gp90pdk0xay7u7fjk72pt6d6pw6uue5n2",
				},
				{
					Type: content.ElementTypeText,
					Text: " link.",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			transformer := content.NewTransformer()

			out, err := transformer.BreakdownAndTransform(testCase.In)
			require.NoError(t, err)

			require.Equal(t, testCase.Out, out)
		})
	}
}
