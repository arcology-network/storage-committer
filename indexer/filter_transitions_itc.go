package indexer

import (
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type ITCTransition struct {
	IPCTransition
	Err error
}

func (this ITCTransition) From(v interfaces.Univalue) interface{} {
	value := IPCTransition{Err: this.Err}.From(v)
	converted := common.IfThenDo1st(value != nil, func() interfaces.Univalue { return value.(interfaces.Univalue) }, nil)
	if converted == nil {
		return nil
	}

	if converted.Value() == nil { // Entry deletion
		return converted
	}

	typed := converted.Value().(interfaces.Type)
	typedNew := typed.New(
		nil,
		typed.CloneDelta(),
		typed.DeltaSign(),
		typed.Min(),
		typed.Max(),
	).(interfaces.Type)

	// typedNew.SetDelta(codec.Clone(typedNew.Delta()))
	converted.SetValue(typedNew) // Reuse the univalue wrapper
	return converted
}
