package indexer

import (
	codec "github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type ITCTransition struct{ IPCTransitions }

func (this ITCTransition) From(v interfaces.Univalue) interface{} {
	value := this.IPCTransitions.From(v)
	converted := common.IfThenDo1st(value != nil, func() interfaces.Univalue { return value.(interfaces.Univalue) }, nil)
	if converted == nil {
		return nil
	}

	if converted.Value() == nil { // Entry deletion
		return converted
	}

	typed := converted.Value().(interfaces.Type)
	typed = typed.New(
		nil,
		codec.Clone(typed.Delta()),
		typed.DeltaSign(),
		typed.Min(),
		typed.Max(),
	).(interfaces.Type)

	converted.SetValue(typed) // Reuse the univalue wrapper
	// converted.Value().(interfaces.Type).SetDelta(codec.Clone(typed.Delta()))
	return converted
}
