package content

import (
	"fmt"

	"github.com/boreq/errors"
)

type Element struct {
	Type ElementType
	Text string
}

var (
	ElementTypeText = ElementType{"text"}
	ElementTypeLink = ElementType{"link"}
)

type ElementType struct {
	s string
}

type Transformer struct {
}

func NewTransformer() *Transformer {
	return &Transformer{}
}

func (t *Transformer) BreakdownAndTransform(content string) ([]Element, error) {
	tokens, err := NewLexer(content).Lex()
	if err != nil {
		return nil, errors.Wrap(err, "error lexing")
	}

	var elements []Element
	for _, token := range tokens {
		switch token.Type {
		case TokenTypeText:
			elements = append(elements, Element{
				Type: ElementTypeText,
				Text: token.Text,
			})
		case TokenTypeLink:
			elements = append(elements, Element{
				Type: ElementTypeLink,
				Text: token.Text,
			})
		case TokenTypeNostrLink:
			elements = append(elements, Element{
				Type: ElementTypeLink,
				Text: fmt.Sprintf("https://njump.me/%s", token.Text),
			})
		default:
			return nil, fmt.Errorf("unknown token '%+v'", token.Type)
		}
	}

	return elements, nil
}
