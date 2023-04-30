package state

import (
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type StateAccess struct {
	encoders []func(ccurlcommon.TypeInterface) []byte
}

func (this StateAccess) New(buffer []ccurlcommon.UnivalueInterface) interface{} {
	// this.encoders = common.Clone(buffer)
	// common.RemoveIf(&this.encoders,
	// 	TransitionFilter{}.ReadOnly,
	// 	TransitionFilter{}.DelNonExist,
	// )

	// common.CastTo(this.buffer, func(v ccurlcommon.UnivalueInterface) codec.Encodeable {
	// 	return common.IfThenDo1st(
	// 		v.Value() != nil &&
	// 			v.DeltaWrites() > 0 &&
	// 			v.Reads() == 0 &&
	// 			v.Writes() == 0 &&
	// 			v.TypeID() != ccurlcommon.Commutative{}.Path(),
	// 		func() codec.Encodeable { return v.Value().(codec.Encodeable) },
	// 		v.Meta().(codec.Encodeable))
	// })
	return this
}

// func (this StateAccess) ToBytes() interface{} {
// 	for i, v := range this.buffer {
// 		if v == nil {
// 			return []byte{} // Deletion
// 		}

// 		v.Value().Encode()

// 	}

// }
