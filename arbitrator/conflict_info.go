package indexer

import (
	"fmt"

	"github.com/arcology-network/common-lib/common"
)

type Conflict struct {
	key   string
	txIDs []uint32
	Err   error
}

func (this Conflict) ToPairs() [][2]uint32 {
	pairs := make([][2]uint32, len(this.txIDs)*(len(this.txIDs)+1)/2-len(this.txIDs))
	for i := 0; i < len(this.txIDs); i++ {
		for j := 0; j < len(this.txIDs); j++ {
			pairs = append(pairs, [2]uint32{this.txIDs[i], this.txIDs[j]})
		}
	}
	return pairs
}

type Conflicts []*Conflict

func (this Conflicts) ToDict() (*map[uint32]uint64, [][2]uint32) {
	dict := make(map[uint32]uint64)
	for _, v := range this {
		for i := 0; i < len(v.txIDs); i++ {
			dict[v.txIDs[i]] += 1
		}
	}
	return &dict, this.ToPairs()
}

func (this Conflicts) Keys() []string {
	keys := make([]string, 0, len(this))
	for _, v := range this {
		keys = append(keys, v.key)
	}
	return keys
}

func (this Conflicts) ToPairs() [][2]uint32 {
	dict := make(map[[2]uint32]int)
	for _, v := range this {
		pairs := v.ToPairs()
		for _, pair := range pairs {
			dict[pair]++
		}
	}
	return common.MapKeys(dict)
}

func (this Conflicts) Print() {
	for _, v := range this {
		fmt.Println(v.key, "      ", v.txIDs)
	}
}
