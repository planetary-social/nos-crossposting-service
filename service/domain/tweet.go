package domain

type Tweet struct {
	text string
}

func NewTweet(text string) Tweet {
	return Tweet{text: text}
}

func (t Tweet) Text() string {
	return t.text
}
