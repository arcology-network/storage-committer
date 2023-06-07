package indexer

import (
	codec "github.com/arcology-network/common-lib/codec"
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
	if typed.IsCommutative() && typed.IsNumeric() { // For the accumulator
		typed = typed.New(
			codec.Clone(typed.Value()),
			codec.Clone(typed.Delta()),
			typed.DeltaSign(),
			typed.Min(),
			typed.Max(),
		).(interfaces.Type)
	} else {
		typed = typed.New(
			nil,
			codec.Clone(typed.Delta()),
			typed.DeltaSign(),
			typed.Min(),
			typed.Max(),
		).(interfaces.Type)
	}

	return converted.New(
		converted.GetUnimeta(),
		typed,
		[]byte{},
	)
}
