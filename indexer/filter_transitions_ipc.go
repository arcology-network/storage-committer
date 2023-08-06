package indexer

import (
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
		nil,
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
