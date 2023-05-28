package univalue

import (
	"github.com/arcology-network/common-lib/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	"github.com/arcology-network/concurrenturl/interfaces"
)

func AccessCodecFilterSet() []func(interfaces.Univalue) interfaces.Univalue {
	return []func(interfaces.Univalue) interfaces.Univalue{
		KeepCumulativeNumeric, // Remove all other transitions, but  Numeric Cumulative variables are neede by the accumulator
	}
}

// Make a deep copy before sending it to other threads
func KeepCumulativeNumeric(v interfaces.Univalue) interfaces.Univalue {
	typedv := common.IfThenDo1st(v != nil && v.Value() != nil, func() interfaces.Type { return v.Value().(interfaces.Type) }, nil)
	if typedv == nil {
		return v
	}

	return (&Univalue{}).New(
		v.GetUnimeta(),
		common.IfThenDo1st(
			typedv != nil && v.DeltaWrites() > 0 && v.Reads() == 0 && v.Writes() == 0 && v.TypeID() != commutative.PATH,
			func() interfaces.Type {
				return typedv.New(nil, typedv.Delta(), typedv.DeltaSign(), typedv.Min(), typedv.Max()).(interfaces.Type)
			},
			nil),
		[]byte{},
		v.GetErrorCode(),
	).(interfaces.Univalue)
}
