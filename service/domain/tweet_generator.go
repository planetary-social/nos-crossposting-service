package domain

import (
	"fmt"
	"strings"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain/content"
)

const noteContentMaxLengthInRunes = 200

const ellipsis = "..."

type TweetGenerator struct {
	transformer *content.Transformer
}

func NewTweetGenerator(transformer *content.Transformer) *TweetGenerator {
	return &TweetGenerator{transformer: transformer}
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

	tweetText, err := g.createText(event)
	if err != nil {
		return nil, errors.Wrap(err, "error creating text")
	}

	return []Tweet{
		NewTweet(tweetText),
	}, nil
}

func (g *TweetGenerator) createText(event Event) (string, error) {
	elements, err := g.transformer.BreakdownAndTransform(event.Content())
	if err != nil {
		return "", errors.Wrap(err, "error transforming")
	}

	var builder strings.Builder
	if err := g.createContent(&builder, elements); err != nil {
		return "", errors.Wrap(err, "error creating content")
	}
	builder.WriteString(fmt.Sprintf("\n\n%s", g.njumpLinkEvent(event)))
	return builder.String(), nil
}

func (g *TweetGenerator) createContent(builder *strings.Builder, elements []content.Element) error {
	for _, element := range elements {
		switch element.Type {
		case content.ElementTypeLink:
			futureTotalLength := builder.Len() + len(element.Text)
			if futureTotalLength > noteContentMaxLengthInRunes {
				builder.WriteString(ellipsis)
				return nil
			}

			builder.WriteString(element.Text)
		case content.ElementTypeText:
			for i, part := range strings.Split(element.Text, " ") {
				futureTotalLength := builder.Len() + len(part)
				if futureTotalLength > noteContentMaxLengthInRunes {
					builder.WriteString(ellipsis)
					return nil
				}

				if i > 0 {
					builder.WriteString(" ")
				}
				builder.WriteString(part)
			}
		default:
			return errors.New("unknown element")
		}
	}

	return nil
}

func (g *TweetGenerator) njumpLinkEvent(event Event) string {
	return fmt.Sprintf("https://njump.me/%s", event.Nevent())
}
