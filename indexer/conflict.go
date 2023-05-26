package indexer

import (
	"github.com/arcology-network/common-lib/common"
)

type Conflict struct {
	key     string
	txIDs   []uint32
	ErrCode uint8
}

type Conflicts []*Conflict

func (this Conflicts) TxIDs() []uint32 {
	txIDs := make([]uint32, 0, len(this))
	for _, v := range this {
		txIDs = append(txIDs, v.txIDs...)
	}
	return common.UniqueInts(txIDs)
}

func (this Conflicts) Keys() []string {
	keys := make([]string, 0, len(this))
	for _, v := range this {
		keys = append(keys, v.key)
	}
	return keys
}
