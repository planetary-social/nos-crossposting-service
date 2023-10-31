package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/boreq/errors"
)

const noteContentMaxLengthInRunes = 200

type TweetGenerator struct {
}

func NewTweetGenerator() *TweetGenerator {
	return &TweetGenerator{}
}

func (g *TweetGenerator) Generate(event Event) ([]Tweet, error) {
	if event.Kind() != EventKindNote {
		return nil, nil
	}

	isReply, err := NoteIsReplyingToOtherEvent(event)
	if err != nil {
		return nil, errors.Wrap(err, "error checking if note is a reply")
	}

	if isReply {
		return nil, nil
	}

	tweetText := g.createText(event)

	return []Tweet{
		NewTweet(tweetText),
	}, nil
}

func (g *TweetGenerator) createText(event Event) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Nostr note made by %s:\n", g.njumpLinkPublicKey(event.PublicKey())))
	builder.WriteString("\n")

	if utf8.RuneCountInString(event.Content()) <= noteContentMaxLengthInRunes {
		builder.WriteString(event.Content())
	} else {
		runeCounter := 0
		for _, rune := range event.Content() {
			builder.WriteRune(rune)
			runeCounter++
			if runeCounter >= noteContentMaxLengthInRunes {
				break
			}
		}
		builder.WriteString("...")
	}

	return builder.String()
}

func (g *TweetGenerator) njumpLinkPublicKey(publicKey PublicKey) string {
	return fmt.Sprintf("https://njump.me/%s", publicKey.Npub())
}
