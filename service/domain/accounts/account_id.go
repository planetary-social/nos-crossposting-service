package accounts

import "github.com/boreq/errors"

type AccountID struct {
	id string
}

func NewAccountID(id string) (AccountID, error) {
	if id == "" {
		return AccountID{}, errors.New("account id can't be an empty string")
	}
	return AccountID{id: id}, nil
}

func MustNewAccountID(id string) AccountID {
	v, err := NewAccountID(id)
	if err != nil {
		panic(err)
	}
	return v
}

func (i AccountID) String() string {
	return i.id
}
