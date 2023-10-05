package sqlite

import (
	"database/sql"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

type PublicKeyRepository struct {
	tx *sql.Tx
}

func NewPublicKeyRepository(tx *sql.Tx) (*PublicKeyRepository, error) {
	return &PublicKeyRepository{
		tx: tx,
	}, nil
}

func (m *PublicKeyRepository) Save(linkedPublicKey *domain.LinkedPublicKey) error {
	_, err := m.tx.Exec(`
	INSERT OR IGNORE INTO public_keys(account_id, public_key, created_at)
	VALUES($1, $2, $3)`,
		linkedPublicKey.AccountID().String(),
		linkedPublicKey.PublicKey().Hex(),
		linkedPublicKey.CreatedAt().Unix(),
	)
	if err != nil {
		return errors.Wrap(err, "error executing the insert query")
	}

	return nil
}
