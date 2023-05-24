package indexer

type Conflict struct {
	key   string
	txIDs []uint32
	err   *error
}

type Conflicts []Conflict

func (this Conflicts) IDs() []uint32 {
	txIDs := []uint32{}
	for _, v := range this {
		txIDs = append(txIDs, v.txIDs...)
	}
	return txIDs
}

func (this Conflicts) Keys() []string {
	keys := []string{}
	for _, v := range this {
		keys = append(keys, v.key)
	}
	return keys
}
