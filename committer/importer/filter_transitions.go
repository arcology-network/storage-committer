package importer

import (
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/storage-committer/commutative"
	"github.com/arcology-network/storage-committer/interfaces"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

// IPTransition stands for intra-process transition. It is used to filter out the fields that are not needed in inter-thread transitions to save
// time spent on encoding and decoding.
type IPTransition struct {
	*univalue.Univalue
	Err error
}

func (this IPTransition) From(v *univalue.Univalue) *univalue.Univalue {
	if v == nil ||
		v.IsReadOnly() ||
		(v.Value() == nil && !v.Preexist()) { // Deletion of an non-existing entry or a read-only entry
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
		common.IfThen(!v.Value().(interfaces.Type).IsCommutative() || common.IsType[*commutative.Path](v.Value()),
			nil,
			v.Value().(interfaces.Type).Value()), // Keep Non-path commutative variables (u256, u64) only
		typed.Delta(),
		typed.DeltaSign(),
		typed.Min(),
		typed.Max(),
	).(interfaces.Type)

	return v.New(
		&v.Property,
		typed,
		[]byte{},
	)
}

// ITT stands for inter-thread transition. It is used to filter out the fields that are not needed in inter-thread transitions to save
// time spent on encoding and decoding, which is only needed in inter-process scenarios.
type ITTransition struct {
	IPTransition
	Err error
}

func (this ITTransition) From(v *univalue.Univalue) *univalue.Univalue {
	unival := IPTransition{Err: this.Err}.From(v)

	// if unival == nil { // Entry deletion
	// 	return unival
	// }

	// converted := common.IfThenDo1st(value != nil, func() *univalue.Univalue { return value.(*univalue.Univalue) }, nil)
	// if converted == nil {
	// 	return nil
	// }

	// The unival is nil when either of the following conditions is met:
	// 1. The unival represents a read-only entry.
	// 2. The unival represents an attempt to delete a non-existing entry.
	if unival == nil || unival.Value() == nil { // Entry deletion
		return unival
	}

	typed := unival.Value().(interfaces.Type) // Get the typed value from the unival
	typed.SetDelta(typed.CloneDelta())
	// typedNew := typed.New(
	// 	nil,
	// 	typed.CloneDelta(),
	// 	typed.DeltaSign(),
	// 	typed.Min(),
	// 	typed.Max(),
	// ).(interfaces.Type)

	// typedNew.SetDelta(codec.Clone(typedNew.Delta()))
	// converted.SetValue(typed) // Reuse the univalue wrapper
	return unival
}
