package accounts

type TwitterID struct {
	id int64
}

func NewTwitterID(id int64) TwitterID {
	return TwitterID{id: id}
}

func (i TwitterID) Int64() int64 {
	return i.id
}
