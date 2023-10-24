package sqlite

import (
	"database/sql"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
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

func (m *PublicKeyRepository) Delete(accountID accounts.AccountID, publicKey domain.PublicKey) error {
	_, err := m.tx.Exec(`
DELETE FROM public_keys
WHERE account_id = $1 AND public_key = $2
`,
		accountID.String(),
		publicKey.Hex(),
	)
	if err != nil {
		return errors.Wrap(err, "error executing the delete query")
	}

	return nil
}

func (m *PublicKeyRepository) List() ([]*domain.LinkedPublicKey, error) {
	rows, err := m.tx.Query(`
SELECT account_id, public_key, created_at
FROM public_keys
`,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	return m.readPublicKeys(rows)
}

func (m *PublicKeyRepository) ListByPublicKey(publicKey domain.PublicKey) ([]*domain.LinkedPublicKey, error) {
	rows, err := m.tx.Query(`
SELECT account_id, public_key, created_at
FROM public_keys
WHERE public_key = $1`,
		publicKey.Hex(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	return m.readPublicKeys(rows)
}

func (m *PublicKeyRepository) ListByAccountID(accountID accounts.AccountID) ([]*domain.LinkedPublicKey, error) {
	rows, err := m.tx.Query(`
SELECT account_id, public_key, created_at
FROM public_keys
WHERE account_id = $1`,
		accountID.String(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	return m.readPublicKeys(rows)
}

func (m *PublicKeyRepository) readPublicKeys(rows *sql.Rows) ([]*domain.LinkedPublicKey, error) {
	var results []*domain.LinkedPublicKey
	for rows.Next() {
		result, err := m.readPublicKey(rows)
		if err != nil {
			return nil, errors.Wrap(err, "error reading public keys")
		}
		results = append(results, result)
	}
	return results, nil
}

func (m *PublicKeyRepository) readPublicKey(row *sql.Rows) (*domain.LinkedPublicKey, error) {
	var accountIDTmp string
	var publicKeyTmp string
	var createdAtTmp int64

	if err := row.Scan(&accountIDTmp, &publicKeyTmp, &createdAtTmp); err != nil {
		return nil, errors.Wrap(err, "error reading the row")
	}

	accountID, err := accounts.NewAccountID(accountIDTmp)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the account id")
	}

	publicKey, err := domain.NewPublicKeyFromHex(publicKeyTmp)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the public key")
	}

	createdAt := time.Unix(createdAtTmp, 0)

	return domain.NewLinkedPublicKey(accountID, publicKey, createdAt)
}
