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

	builder.WriteString(fmt.Sprintf("\n\n%s", g.njumpLinkEvent(event)))

	return builder.String()
}

func (g *TweetGenerator) njumpLinkEvent(event Event) string {
	return fmt.Sprintf("https://njump.me/%s", event.Nevent())
}
