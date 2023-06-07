package indexer

import (
	"bytes"
	"sort"

	codec "github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/commutative"
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

// // func RemoveReadOnly(unival interfaces.Univalue) interfaces.Univalue {
// // 	return common.IfThen(unival != nil && unival.IsReadOnly(), nil, unival)
// // }

func ExtractDeltaForEncoding(unival interfaces.Univalue) interfaces.Univalue {
	if unival == nil {
		return unival
	}

	var v interface{}
	if unival.Value() != nil {
		value := unival.Value().(interfaces.Type)
		if unival.DeltaWrites() > 0 && unival.TypeID() != commutative.PATH { // commutative but not meta, for the accumulator
			v = unival.Value().(interfaces.Type).New(
				value.Value(), // The value is needed for the accumulator, otherwise it has no access to its original state
				common.IfThenDo1st(value.Delta() != nil, func() interface{} { return value.Delta().(codec.Encodable).Clone() }, nil),
				value.DeltaSign(),
				common.IfThenDo1st(value.Min() != nil, func() interface{} { return value.Min().(codec.Encodable).Clone() }, nil),
				common.IfThenDo1st(value.Max() != nil, func() interface{} { return value.Max().(codec.Encodable).Clone() }, nil),
			)

		} else {
			v = unival.Value().(interfaces.Type).New(nil, value.Delta(), value.DeltaSign(), nil, nil)
		}
	}
	// meta := unival.GetUnimeta().(*Unimeta)
	return unival.New(
		unival.GetUnimeta(),
		v,
		[]byte{},
	).(interfaces.Univalue)
}

// func IsReadOnly(v interfaces.Univalue) bool    { return v != nil && v.IsReadOnly() }
// func IsDelNonExist(v interfaces.Univalue) bool { return v != nil && v.Value() == nil && !v.Preexist() }
// func IsNonce(unival interfaces.Univalue) bool  { return strings.HasSuffix(*unival.GetPath(), "/nonce") }
// func IsBalance(v interfaces.Univalue) bool     { return strings.HasSuffix(*v.GetPath(), "/balance") }

// // func RemoveNonce(unival interfaces.Univalue) interfaces.Univalue {
// // 	return common.IfThen(
// // 		unival != nil && len(*unival.GetPath()) > len("/nonce") && (*unival.GetPath())[len(*unival.GetPath())-len("/nonce"):] == "/nonce",
// // 		nil,
// // 		unival,
// // 	)
// // }

// func CloneValue(unival interfaces.Univalue) interfaces.Univalue {
// 	return common.IfThenDo1st(
// 		unival != nil,
// 		func() interfaces.Univalue { return unival.Clone().(interfaces.Univalue) }, unival)
// }

// func TransitionCodecFilterSet() []interfaces.Univalue {
// 	return []interfaces.Univalue{
// 		this.From()
// 		// RemoveReadOnly,
// 		// DelNonExist,
// 		// ExtractDeltaForEncoding, //
// 	}
// }
