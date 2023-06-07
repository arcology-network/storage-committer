package indexer

import (
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type IPCAccess struct {
	interfaces.Univalue
	Status uint8
}

func (this IPCAccess) From(v interfaces.Univalue) interface{} {
	if this.Status != ccurlcommon.SUCCESSFUL {
		return nil
	}

	if v.Value() == nil {
		return v
	}

	value := v.Value().(interfaces.Type)
	return v.New(
		v.GetUnimeta(),
		common.IfThen(value.IsCommutative() && value.IsNumeric(), value, nil), // commutative but not meta, for the accumulator
		[]byte{},
	)
}
