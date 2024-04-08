package statestore

import (
	"bytes"
	"sort"

	univalue "github.com/arcology-network/storage-committer/univalue"
)

func Sorter(univals []*univalue.Univalue) []*univalue.Univalue {
	sort.SliceStable(univals, func(i, j int) bool {
		lhs := (*(univals[i].GetPath()))
		rhs := (*(univals[j].GetPath()))
		return bytes.Compare([]byte(lhs)[:], []byte(rhs)[:]) < 0
	})
	return univals
}
