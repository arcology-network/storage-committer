package indexer

import (
	"bytes"
	"sort"

	"github.com/arcology-network/concurrenturl/interfaces"
)

func Sorter(univals []interfaces.Univalue) []interfaces.Univalue {
	sort.SliceStable(univals, func(i, j int) bool {
		lhs := (*(univals[i].GetPath()))
		rhs := (*(univals[j].GetPath()))
		return bytes.Compare([]byte(lhs)[:], []byte(rhs)[:]) < 0
	})
	return univals
}
