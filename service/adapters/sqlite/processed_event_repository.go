package sqlite

import (
	"database/sql"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type ProcessedEventRepository struct {
	tx *sql.Tx
}

func NewProcessedEventRepository(tx *sql.Tx) (*ProcessedEventRepository, error) {
	return &ProcessedEventRepository{
		tx: tx,
	}, nil
}

func (m *ProcessedEventRepository) Save(eventID domain.EventId, twitterID accounts.TwitterID) error {
	_, err := m.tx.Exec(`
	INSERT OR IGNORE INTO processed_events(twitter_id, event_id)
	VALUES($1, $2)`,
		twitterID.Int64(),
		eventID.Hex(),
	)
	if err != nil {
		return errors.Wrap(err, "error executing the insert query")
	}

	return nil
}

func (m *ProcessedEventRepository) WasProcessed(eventID domain.EventId, twitterID accounts.TwitterID) (bool, error) {
	row := m.tx.QueryRow(`
SELECT twitter_id, event_id
FROM processed_events
WHERE twitter_id = $1 AND event_id = $2`,
		twitterID.Int64(),
		eventID.Hex(),
	)

	var twitterIDTmp int64
	var eventIDTmp string

	if err := row.Scan(&twitterIDTmp, &eventIDTmp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, errors.Wrap(err, "query error")
	}
	return true, nil
}
