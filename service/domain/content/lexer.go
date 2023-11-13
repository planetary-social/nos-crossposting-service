package content

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/boreq/errors"
)

const (
	httpColonSlashSlash  = "http://"
	httpsColonSlashSlash = "https://"
	nostrColon           = "nostr:"
	nevent               = "nevent"
	npub                 = "npub"
	note                 = "note"
)

type Token struct {
	Type TokenType
	Text string
}

var (
	TokenTypeText      = TokenType{"text"}
	TokenTypeLink      = TokenType{"link"}
	TokenTypeNostrLink = TokenType{"nostrLink"}
)

type TokenType struct {
	s string
}

type Lexer struct {
	in     string
	out    string
	tokens []Token
}

func NewLexer(s string) *Lexer {
	return &Lexer{in: s}
}

func (l *Lexer) Lex() ([]Token, error) {
	if !utf8.ValidString(l.in) {
		return nil, errors.New("invalid utf-8")
	}

	var state stateFn = stateText

	for state != nil {
		newState, err := state(l)
		if err != nil {
			return nil, errors.Wrap(err, "state fn error")
		}

		state = newState
	}

	return l.mergeConsecutiveTexts(l.tokens), nil
}

func (l *Lexer) next() (rune, bool) {
	r, size := utf8.DecodeRuneInString(l.in)
	if r == utf8.RuneError && size == 0 {
		return 0, false
	}
	if r == utf8.RuneError && size == 1 {
		panic("invalid utf-8") // checked in constructor
	}
	l.in = l.in[size:]
	l.out += string(r)
	return r, true
}

func (l *Lexer) back() error {
	r, size := utf8.DecodeLastRuneInString(l.out)
	if r == utf8.RuneError && size == 0 {
		return errors.New("empty out")
	}
	if r == utf8.RuneError && size == 1 {
		return errors.New("this should be impossible, out is malformed")
	}
	l.out = l.out[:len(l.out)-size]
	l.in = string(r) + l.in
	return nil
}

func (l *Lexer) backN(n int) error {
	for i := 0; i < n; i++ {
		if err := l.back(); err != nil {
			return errors.Wrapf(err, "error calling back for index '%d'", i)
		}
	}
	return nil
}

func (l *Lexer) tryOrBack(s string) bool {
	counter := 0
	for _, expectedR := range s {
		counter++
		nextR, ok := l.next()
		if !ok || nextR != expectedR {
			if err := l.backN(counter); err != nil {
				panic(err) // we just called next `counter` times so if there is a bug here it would have to be insanely silly
			}
			return false
		}
	}
	return true
}

func (l *Lexer) comesNext(s string) bool {
	counter := 0

	defer func() {
		if err := l.backN(counter); err != nil {
			panic(err) // we just called next `counter` times so if there is a bug here it would have to be insanely silly
		}
	}()

	for _, expectedR := range s {
		nextR, ok := l.next()
		if ok {
			counter++
		}
		if !ok || nextR != expectedR {
			return false
		}
	}
	return true

}

func (l *Lexer) emit(typ TokenType) {
	if l.out != "" {
		l.tokens = append(l.tokens,
			Token{
				Type: typ,
				Text: l.out,
			},
		)
		l.out = ""
	}
}

func (l *Lexer) peek() (rune, bool) {
	r, ok := l.next()
	if ok {
		if err := l.back(); err != nil {
			panic(err)
		}
	}
	return r, ok
}

func (l *Lexer) destroyPrefix(prefix string) {
	l.out = strings.TrimPrefix(l.out, prefix)
}

// honestly I feel like this is easier than debugging the lexer so that it
// doesn't make those mistakes
func (l *Lexer) mergeConsecutiveTexts(tokens []Token) []Token {
	var result []Token
	for _, token := range tokens {
		if len(result) > 0 && token.Type == TokenTypeText && result[len(result)-1].Type == TokenTypeText {
			result[len(result)-1].Text += token.Text
		} else {
			result = append(result, token)
		}
	}

	return result

}

type stateFn func(l *Lexer) (stateFn, error)

func stateText(l *Lexer) (stateFn, error) {
	for {
		if l.comesNext(httpColonSlashSlash) || l.comesNext(httpsColonSlashSlash) {
			l.emit(TokenTypeText)
			return stateLinkProtocol, nil
		}

		if l.comesNext(nostrColon) {
			l.emit(TokenTypeText)
			return stateNostrLinkProtocol, nil
		}

		if l.comesNext(nevent) || l.comesNext(npub) || l.comesNext(note) {
			l.emit(TokenTypeText)
			return stateNostrLinkType, nil
		}

		_, ok := l.next()
		if !ok {
			l.emit(TokenTypeText)
			return nil, nil
		}
	}
}

func stateLinkProtocol(l *Lexer) (stateFn, error) {
	if !l.tryOrBack(httpColonSlashSlash) && !l.tryOrBack(httpsColonSlashSlash) {
		return nil, errors.New("where did the protocol go?")
	}

	return stateLinkAddress, nil
}

func stateLinkAddress(l *Lexer) (stateFn, error) {
	counter := 0
	for {
		r, ok := l.next()
		if !ok {
			if counter == 0 {
				l.emit(TokenTypeText)
			} else {
				l.emit(TokenTypeLink)
			}
			return nil, nil
		}

		counter++

		if isValidLinkCharacterExcludingDot(r) {
			continue
		}

		switch r {
		case '.':
			nextR, ok := l.peek()
			if !ok {
				continue
			}

			if !isValidLinkCharacterExcludingDot(nextR) {
				if err := l.back(); err != nil {
					return nil, errors.New("where did the dot go?")
				}
				l.emit(TokenTypeLink)
				return stateText, nil
			}
		default:
			if err := l.back(); err != nil {
				return nil, errors.New("we just went forward but we can't go back?")
			}
			l.emit(TokenTypeLink)
			return stateText, nil
		}
	}
}

func stateNostrLinkProtocol(l *Lexer) (stateFn, error) {
	if !l.tryOrBack(nostrColon) {
		return nil, errors.New("where did 'nostr:' go?")
	}

	return stateNostrLinkType, nil
}

func stateNostrLinkType(l *Lexer) (stateFn, error) {
	if !l.tryOrBack(nevent) && !l.tryOrBack(npub) && !l.tryOrBack(note) {
		return stateText, nil
	}

	return stateNostrLinkData, nil
}

func stateNostrLinkData(l *Lexer) (stateFn, error) {
	counter := 0
	for {
		r, ok := l.next()
		if !ok || !isBech32(r) {
			if ok {
				if err := l.back(); err != nil {
					return nil, errors.Wrap(err, "we just consumed a rune?")
				}
			}

			if counter == 0 {
				l.emit(TokenTypeText)
			} else {
				l.destroyPrefix(nostrColon)
				l.emit(TokenTypeNostrLink)
			}

			return stateText, nil
		}

		counter++
	}
}

func isValidLinkCharacterExcludingDot(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '%' || r == '/'
}

func isBech32(r rune) bool {
	r = unicode.ToLower(r)
	if r == 'b' || r == 'i' || r == 'o' {
		return false
	}
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}
