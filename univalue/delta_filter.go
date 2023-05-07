package univalue

import (
	"bytes"
	"sort"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
)

func ReadOnly(unival ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	return common.IfThen(unival != nil && unival.IsReadOnly(), nil, unival)
}

func DelNonExist(unival ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	return common.IfThen(unival != nil && unival.Value() == nil && !unival.Preexist(), nil, unival)
}

func Sorter(univals []ccurlcommon.UnivalueInterface) []ccurlcommon.UnivalueInterface {
	sort.SliceStable(univals, func(i, j int) bool {
		lhs := (*(univals[i].GetPath()))
		rhs := (*(univals[j].GetPath()))
		return bytes.Compare([]byte(lhs)[:], []byte(rhs)[:]) < 0
	})
	return univals
}

func ExtractDelta(unival ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	if unival == nil {
		return unival
	}

	var v interface{}
	if unival.Value() != nil {
		value := unival.Value().(ccurlcommon.TypeInterface)
		if !unival.Preexist() || (unival.DeltaWrites() > 0 && unival.TypeID() != commutative.PATH) { // commutative but not meta, for the accumulator
			v = unival.Value().(ccurlcommon.TypeInterface).New(
				nil,
				common.IfThenDo1st(value.Delta() != nil, func() interface{} { return value.Delta().(codec.Encodable).Clone() }, nil),
				value.Sign(),
				common.IfThenDo1st(value.Min() != nil, func() interface{} { return value.Min().(codec.Encodable).Clone() }, nil),
				common.IfThenDo1st(value.Max() != nil, func() interface{} { return value.Max().(codec.Encodable).Clone() }, nil),
			)

		} else {
			v = unival.Value().(ccurlcommon.TypeInterface).New(nil, value.Delta(), value.Sign(), nil, nil)
		}
	}

	return &Univalue{
		unival.GetUnimeta().(Unimeta),
		v,
		[]byte{},
	}
}

func TransitionFilters() []func(ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	return []func(ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface{
		ReadOnly,
		DelNonExist,
		ExtractDelta, //
	}
}
