package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type ProcessedEventRepository struct {
}

func NewProcessedEventRepository() (*ProcessedEventRepository, error) {
	return &ProcessedEventRepository{}, nil
}

func (m *ProcessedEventRepository) Save(eventID domain.EventId, twitterID accounts.TwitterID) error {
	return errors.New("not implemented")
}

func (m *ProcessedEventRepository) WasProcessed(eventID domain.EventId, twitterID accounts.TwitterID) (bool, error) {
	return false, errors.New("not implemented")
}
