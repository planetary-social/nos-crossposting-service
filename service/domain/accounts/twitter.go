package accounts

import "github.com/boreq/errors"

type TwitterID struct {
	id int64
}

func NewTwitterID(id int64) TwitterID {
	return TwitterID{id: id}
}

func (i TwitterID) Int64() int64 {
	return i.id
}

type TwitterUserAccessToken struct {
	s string
}

func NewTwitterUserAccessToken(s string) (TwitterUserAccessToken, error) {
	if s == "" {
		return TwitterUserAccessToken{}, errors.New("user access token can't be empty")
	}
	return TwitterUserAccessToken{s: s}, nil
}

func (t TwitterUserAccessToken) String() string {
	return t.s
}

type TwitterUserAccessSecret struct {
	s string
}

func NewTwitterUserAccessSecret(s string) (TwitterUserAccessSecret, error) {
	if s == "" {
		return TwitterUserAccessSecret{}, errors.New("user access secret can't be empty")
	}
	return TwitterUserAccessSecret{s: s}, nil
}

func (t TwitterUserAccessSecret) String() string {
	return t.s
}

type TwitterUserTokens struct {
	accountID    AccountID
	accessToken  TwitterUserAccessToken
	accessSecret TwitterUserAccessSecret
}

func NewTwitterUserTokens(
	accountID AccountID,
	accessToken TwitterUserAccessToken,
	accessSecret TwitterUserAccessSecret,
) *TwitterUserTokens {
	return &TwitterUserTokens{
		accountID:    accountID,
		accessToken:  accessToken,
		accessSecret: accessSecret,
	}
}

func (t TwitterUserTokens) AccountID() AccountID {
	return t.accountID
}

func (t TwitterUserTokens) AccessToken() TwitterUserAccessToken {
	return t.accessToken
}

func (t TwitterUserTokens) AccessSecret() TwitterUserAccessSecret {
	return t.accessSecret
}
