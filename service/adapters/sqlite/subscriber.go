package sqlite

import (
	"context"
	"database/sql"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type Subscriber struct {
	pubsub *PubSub
	db     *sql.DB
}

func NewSubscriber(
	pubsub *PubSub,
	db *sql.DB,
) *Subscriber {
	return &Subscriber{
		pubsub: pubsub,
		db:     db,
	}
}

func (s *Subscriber) SubscribeToTweetCreated(ctx context.Context) <-chan *ReceivedMessage {
	return s.pubsub.Subscribe(ctx, TweetCreatedTopic)
}

func (s *Subscriber) TweetCreatedQueueLength(ctx context.Context) (int, error) {
	return s.pubsub.QueueLength(TweetCreatedTopic)
}

func (s *Subscriber) TweetCreatedAnalysis(ctx context.Context) (TweetCreatedAnalysis, error) {
	analysis := TweetCreatedAnalysis{
		TweetsPerAccountID: make(map[accounts.AccountID]int),
	}

	rows, err := s.db.Query(
		"SELECT json_extract(payload, '$.accountID') as accountID, COUNT(*) FROM pubsub WHERE topic = ? GROUP BY accountID",
		TweetCreatedTopic,
	)
	if err != nil {
		return TweetCreatedAnalysis{}, errors.Wrap(err, "query error")
	}

	for rows.Next() {
		var (
			accountIDPrimitive string
			count              int
		)
		if err := rows.Scan(&accountIDPrimitive, &count); err != nil {
			return TweetCreatedAnalysis{}, errors.Wrap(err, "scan error")
		}

		accountID, err := accounts.NewAccountID(accountIDPrimitive)
		if err != nil {
			return TweetCreatedAnalysis{}, errors.Wrap(err, "error creating account id")
		}

		analysis.TweetsPerAccountID[accountID] = count
	}

	return analysis, nil
}

type TweetCreatedAnalysis struct {
	TweetsPerAccountID map[accounts.AccountID]int
}
