package sqlite

import (
	"database/sql"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type AccountRepository struct {
	tx *sql.Tx
}

func NewAccountRepository(tx *sql.Tx) (*AccountRepository, error) {
	return &AccountRepository{
		tx: tx,
	}, nil
}

func (m *AccountRepository) GetByTwitterID(twitterID accounts.TwitterID) (*accounts.Account, error) {
	result := m.tx.QueryRow(`
SELECT account_id, twitter_id
FROM accounts
WHERE twitter_id=$1`,
		twitterID.Int64(),
	)

	return m.readAccount(result)
}

func (m *AccountRepository) GetByAccountID(accountID accounts.AccountID) (*accounts.Account, error) {
	result := m.tx.QueryRow(`
SELECT account_id, twitter_id
FROM accounts
WHERE account_id=$1`,
		accountID.String(),
	)

	return m.readAccount(result)
}

func (m *AccountRepository) Save(account *accounts.Account) error {
	_, err := m.tx.Exec(`
INSERT INTO accounts(account_id, twitter_id)
VALUES('$1','$2')
ON CONFLICT(account_id) DO UPDATE SET
  twitter_id=excluded.twitter_id`,
		account.AccountID().String(),
		account.TwitterID().Int64(),
	)
	if err != nil {
		return errors.Wrap(err, "error executing the insert query")
	}

	return nil
}

func (m *AccountRepository) readAccount(result *sql.Row) (*accounts.Account, error) {
	var accountIDtmp string
	var twitterIDtmp int64

	if err := result.Scan(&accountIDtmp, &twitterIDtmp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, app.ErrAccountDoesNotExist
		}
		return nil, errors.Wrap(err, "error reading the row")
	}

	accountID, err := accounts.NewAccountID(accountIDtmp)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the account id")
	}

	twitterID := accounts.NewTwitterID(twitterIDtmp)

	return accounts.NewAccount(accountID, twitterID)
}
