package accounts

type Account struct {
	accountID AccountID
	twitterID TwitterID
}

func NewAccount(accountID AccountID, twitterID TwitterID) (*Account, error) {
	return &Account{accountID: accountID, twitterID: twitterID}, nil
}

func (a Account) AccountID() AccountID {
	return a.accountID
}

func (a Account) TwitterID() TwitterID {
	return a.twitterID
}
