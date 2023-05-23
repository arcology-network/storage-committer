package univalue

import (
	"bytes"
	"sort"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
)

func Sorter(univals []ccurlcommon.UnivalueInterface) []ccurlcommon.UnivalueInterface {
	sort.SliceStable(univals, func(i, j int) bool {
		lhs := (*(univals[i].GetPath()))
		rhs := (*(univals[j].GetPath()))
		return bytes.Compare([]byte(lhs)[:], []byte(rhs)[:]) < 0
	})
	return univals
}

func RemoveReadOnly(unival ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	return common.IfThen(unival != nil && unival.IsReadOnly(), nil, unival)
}

func DelNonExist(unival ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	return common.IfThen(unival != nil && unival.Value() == nil && !unival.Preexist(), nil, unival)
}

func ExtractDeltaForEncoding(unival ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	if unival == nil {
		return unival
	}

	var v interface{}
	if unival.Value() != nil {
		value := unival.Value().(ccurlcommon.TypeInterface)
		if unival.DeltaWrites() > 0 && unival.TypeID() != commutative.PATH { // commutative but not meta, for the accumulator
			v = unival.Value().(ccurlcommon.TypeInterface).New(
				value.Value(), // The value is needed for the accumulator, otherwise it has no access to its original state
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

func RemoveNonce(unival ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	return common.IfThen(
		unival != nil && len(*unival.GetPath()) > len("/nonce") && (*unival.GetPath())[len(*unival.GetPath())-len("/nonce"):] == "/nonce",
		nil,
		unival,
	)
}

func CloneValue(unival ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	return common.IfThenDo1st(
		unival != nil,
		func() ccurlcommon.UnivalueInterface { return unival.Clone().(ccurlcommon.UnivalueInterface) }, unival)
}

func TransitionCodecFilterSet() []func(ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	return []func(ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface{
		RemoveReadOnly,
		DelNonExist,
		ExtractDeltaForEncoding, //
	}
}
