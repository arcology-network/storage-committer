package univalue

import (
	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	commutative "github.com/arcology-network/concurrenturl/v2/commutative"
)

func ReadOnly(unival ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	return common.IfThen(unival != nil && unival.IsReadOnly(), nil, unival)
}

func DelNonExist(unival ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	return common.IfThen(unival != nil && unival.Value() == nil && !unival.Preexist(), nil, unival)
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
