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

// func ExtractDeltaForEncoding(unival interfaces.Univalue) interfaces.Univalue {
// 	if unival == nil {
// 		return unival
// 	}

// 	var v interface{}
// 	if unival.Value() != nil {
// 		value := unival.Value().(interfaces.Type)
// 		if unival.DeltaWrites() > 0 && unival.TypeID() != commutative.PATH { // commutative but not meta, for the accumulator
// 			v = unival.Value().(interfaces.Type).New(
// 				value.Value(), // The value is needed for the accumulator, otherwise it has no access to its original state
// 				common.IfThenDo1st(value.Delta() != nil, func() interface{} { return value.Delta().(codec.Encodable).Clone() }, nil),
// 				value.DeltaSign(),
// 				common.IfThenDo1st(value.Min() != nil, func() interface{} { return value.Min().(codec.Encodable).Clone() }, nil),
// 				common.IfThenDo1st(value.Max() != nil, func() interface{} { return value.Max().(codec.Encodable).Clone() }, nil),
// 			)

// 		} else {
// 			v = unival.Value().(interfaces.Type).New(nil, value.Delta(), value.DeltaSign(), nil, nil)
// 		}
// 	}
// 	// meta := unival.GetUnimeta().(*Unimeta)
// 	return unival.New(
// 		unival.GetUnimeta(),
// 		v,
// 		[]byte{},
// 	).(interfaces.Univalue)
// }
