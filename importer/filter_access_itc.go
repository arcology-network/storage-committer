package indexer

import (
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type ITCAccess struct{ IPCAccess }

func (this ITCAccess) From(v interfaces.Univalue) interface{} {
	value := this.IPCAccess.From(v)
	converted := common.IfThenDo1st(value != nil, func() interfaces.Univalue { return value.(interfaces.Univalue) }, nil)
	if converted == nil {
		return nil
	}

	if converted.Value() == nil { // Entry deletion
		return converted
	}

	typed := converted.Value().(interfaces.Type)
	newv := typed.New(
		nil,
		typed.Delta(),
		typed.DeltaSign(),
		typed.Min(),
		typed.Max(),
	).(interfaces.Type)

	if typed.IsCommutative() && typed.IsNumeric() { // For the accumulator, commutative u64 & U256
		newv.SetValue(typed.Value())
	}

	return converted.New(
		converted.GetUnimeta(),
		typed,
		[]byte{},
	)
}
