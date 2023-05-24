package univalue

import (
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
)

func AccessCodecFilterSet() []func(ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	return []func(ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface{
		KeepCumulativeNumeric, // Remove all other transitions, but  Numeric Cumulative variables are neede by the accumulator
	}
}

// Make a deep copy before sending it to other threads
func KeepCumulativeNumeric(v ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface {
	typedv := common.IfThenDo1st(v != nil && v.Value() != nil, func() ccurlcommon.TypeInterface { return v.Value().(ccurlcommon.TypeInterface) }, nil)
	if typedv == nil {
		return v
	}

	return (&Univalue{}).New(
		v.GetUnimeta(),
		common.IfThenDo1st(
			typedv != nil && v.DeltaWrites() > 0 && v.Reads() == 0 && v.Writes() == 0 && v.TypeID() != commutative.PATH,
			func() ccurlcommon.TypeInterface {
				return typedv.New(nil, typedv.Delta(), typedv.DeltaSign(), typedv.Min(), typedv.Max()).(ccurlcommon.TypeInterface)
			},
			nil),
		[]byte{},
	).(ccurlcommon.UnivalueInterface)
}
