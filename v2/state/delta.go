package state

import (
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	commutative "github.com/arcology-network/concurrenturl/v2/commutative"
)

type StateDelta struct {
	buffer []ccurlcommon.UnivalueInterface
}

func (this StateDelta) New(buffer []ccurlcommon.UnivalueInterface) StateDelta {
	this.buffer = common.Clone(buffer)
	common.RemoveIf(&this.buffer,
		TransitionFilter{}.ReadOnly,
		TransitionFilter{}.DelNonExist,
	)

	common.CastTo(this.buffer, func(v ccurlcommon.UnivalueInterface) codec.Encodeable {
		return common.IfThenDo1st(
			v.Value() != nil &&
				v.IsReadOnly() && v.DeltaWrites() > 0 && v.TypeID() != commutative.PATH,
			func() codec.Encodeable { return v.Value().(codec.Encodeable) },
			v.Meta().(codec.Encodeable))
	})
	return this
}
