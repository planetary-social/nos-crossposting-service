package mocks

import (
	"github.com/planetary-social/nos-crossposting-service/service/app"
)

type Publisher struct {
}

func NewPublisher() *Publisher {
	return &Publisher{}
}

func (p *Publisher) PublishTweetCreated(event app.TweetCreatedEvent) error {
	return nil
}
