package sqlite

import (
	"context"
)

type Subscriber struct {
	pubsub *PubSub
}

func NewSubscriber(
	pubsub *PubSub,
) *Subscriber {
	return &Subscriber{
		pubsub: pubsub,
	}
}

func (s *Subscriber) SubscribeToTweetCreated(ctx context.Context) <-chan *ReceivedMessage {
	return s.pubsub.Subscribe(ctx, TweetCreatedTopic)
}

func (s *Subscriber) TweetCreatedQueueLength(ctx context.Context) (int, error) {
	return s.pubsub.QueueLength(TweetCreatedTopic)
}
