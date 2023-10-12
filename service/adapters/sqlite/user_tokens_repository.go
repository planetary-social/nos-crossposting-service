package sqlite

import (
	"database/sql"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
)

type UserTokensRepository struct {
	tx *sql.Tx
}

func NewUserTokensRepository(tx *sql.Tx) (*UserTokensRepository, error) {
	return &UserTokensRepository{
		tx: tx,
	}, nil
}

func (m *UserTokensRepository) Save(userTokens *accounts.TwitterUserTokens) error {
	_, err := m.tx.Exec(`
	INSERT OR IGNORE INTO user_tokens(account_id, access_token, access_secret)
	VALUES($1, $2, $3)
	ON CONFLICT(account_id) DO UPDATE SET
	  access_token=excluded.access_token,
	  access_secret=excluded.access_secret`,
		userTokens.AccountID().String(),
		userTokens.AccessToken().String(),
		userTokens.AccessSecret().String(),
	)
	if err != nil {
		return errors.Wrap(err, "error executing the insert query")
	}

	return nil
}

func (m *UserTokensRepository) Get(id accounts.AccountID) (*accounts.TwitterUserTokens, error) {
	result := m.tx.QueryRow(`
SELECT account_id, access_token, access_secret
FROM user_tokens
WHERE account_id=$1`,
		id.String(),
	)

	return m.readUserTokens(result)
}

func (m *UserTokensRepository) readUserTokens(result *sql.Row) (*accounts.TwitterUserTokens, error) {
	var accountIDTmp string
	var accessTokenTmp string
	var accessSecretTmp string

	if err := result.Scan(&accountIDTmp, &accessTokenTmp, &accessSecretTmp); err != nil {
		return nil, errors.Wrap(err, "error reading the row")
	}

	accountID, err := accounts.NewAccountID(accountIDTmp)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the account id")
	}

	accessToken, err := accounts.NewTwitterUserAccessToken(accessTokenTmp)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the access token")
	}

	accessSecret, err := accounts.NewTwitterUserAccessSecret(accessSecretTmp)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the access secret")
	}

	return accounts.NewTwitterUserTokens(accountID, accessToken, accessSecret), nil
}
