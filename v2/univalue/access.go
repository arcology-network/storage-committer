package univalue

import (
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	commutative "github.com/arcology-network/concurrenturl/v2/commutative"
	"github.com/arcology-network/concurrenturl/v2/state"
)

type AccessEncoder Univalues

func (this AccessEncoder) Encode() []byte {
	buffer := common.Clone(this)
	common.RemoveIf(&buffer,
		state.TransitionFilter{}.ReadOnly,
		state.TransitionFilter{}.DelNonExist,
	)

	buffers := make([][]byte, len(this))

	for i, v := range this {
		var flags []interface{}
		if v.DeltaWrites() > 0 && v.Reads() == 0 && v.Writes() == 0 && v.TypeID() != commutative.PATH {
			flags = []interface{}{true, true, true, true}
		} else {
			flags = []interface{}{false, false, false, false}
		}
		buffers[i] = v.Encode(flags...)
	}
	return codec.Byteset(buffers).Encode()
}
