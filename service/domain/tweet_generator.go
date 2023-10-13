package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"
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

	tweetText := g.createText(event)

	return []Tweet{
		NewTweet(tweetText),
	}, nil
}

func (g *TweetGenerator) createText(event Event) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Nostr note made by %s:\n", event.PublicKey().Npub()))
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
