package indexer

import (
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/commutative"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type IPCTransition struct {
	interfaces.Univalue
	Err error
}

func (this IPCTransition) From(v interfaces.Univalue) interface{} {
	if v == nil ||
		v.IsReadOnly() ||
		(v.Value() == nil && !v.Preexist()) { // Del Non Exist
		return nil
	}

	if v.Value() == nil { // Entry deletion
		return v
	}

	if this.Err != nil && !v.Persistent() { // Keep balance and nonce transitions for failed ones.
		return nil
	}

	typed := v.Value().(interfaces.Type)
	typed = typed.New(
		common.IfThen(v.Persistent() || common.IsType[*commutative.Path](v.Value()), nil, v.Value().(interfaces.Type).Value()), //,nil,
		typed.Delta(),
		typed.DeltaSign(),
		typed.Min(),
		typed.Max(),
	).(interfaces.Type)

	return v.New(
		v.GetUnimeta(),
		typed,
		[]byte{},
	)
}
