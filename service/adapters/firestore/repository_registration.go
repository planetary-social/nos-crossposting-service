package firestore

import (
	"cloud.google.com/go/firestore"
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

type RegistrationRepository struct {
	client *firestore.Client
	tx     *firestore.Transaction

	relayRepository     *RelayRepository
	publicKeyRepository *PublicKeyRepository
}

func NewRegistrationRepository(
	client *firestore.Client,
	tx *firestore.Transaction,
	relayRepository *RelayRepository,
	publicKeyRepository *PublicKeyRepository,
) *RegistrationRepository {
	return &RegistrationRepository{
		client:              client,
		tx:                  tx,
		relayRepository:     relayRepository,
		publicKeyRepository: publicKeyRepository,
	}
}

func (r *RegistrationRepository) Save(registration domain.Registration) error {
	if err := r.relayRepository.Save(registration); err != nil {
		return errors.Wrap(err, "error saving under relays")
	}

	if err := r.publicKeyRepository.Save(registration); err != nil {
		return errors.Wrap(err, "error saving under public keys")
	}

	return nil
}
