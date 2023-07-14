package indexer

import "fmt"

type Conflict struct {
	key   string
	txIDs []uint32
	Err   error
}

type Conflicts []*Conflict

func (this Conflicts) ToDict() *map[uint32]uint64 {
	dict := make(map[uint32]uint64)
	for _, v := range this {
		for i := 0; i < len(v.txIDs); i++ {
			dict[v.txIDs[i]] += 1
		}
	}
	return &dict
}

func (this Conflicts) Keys() []string {
	keys := make([]string, 0, len(this))
	for _, v := range this {
		keys = append(keys, v.key)
	}
	return keys
}

func (this Conflicts) Print() {
	for _, v := range this {
		fmt.Println(v.key, "      ", v.txIDs)
	}
}
