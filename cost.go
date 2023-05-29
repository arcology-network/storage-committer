package concurrenturl

import (
	common "github.com/arcology-network/common-lib/common"
	interfaces "github.com/arcology-network/concurrenturl/interfaces"
	// ethparams "github.com/arcology-network/evm/params"
)

const (
	IS_PATH = uint64(3)
)

type Cost struct{}

func (Cost) Reader(univ interfaces.Univalue) uint64 { // Call this before setting the value attribute to nil
	typedv := univ.Value()
	dataSize := common.IfThenDo1st(typedv != nil, func() uint64 { return uint64(typedv.(interfaces.Type).MemSize()) }, 0)

	return common.IfThen(
		univ.IsHotLoaded(),
		common.Max(dataSize/32, 1)*3,
		(dataSize/32)*5000,
	)
}

func (Cost) Writer(key string, v interface{}, writecache interfaces.WriteCache) int64 { // May get refunds sometimes
	committedv := writecache.RetriveShallow(key)
	committedSize := common.IfThenDo1st(committedv != nil, func() uint64 { return uint64(committedv.(interfaces.Type).MemSize()) }, 0)

	dataSize := common.IfThenDo1st(v != nil, func() uint64 { return uint64(v.(interfaces.Type).Size()) }, 0)
	return common.IfThen(
		committedSize > 0,
		common.Max(int64(dataSize/32), 1)*20000,
		common.IfThen(
			int64(dataSize) > int64(committedSize),
			int64(dataSize),
			((int64(committedSize)-int64(dataSize))/32)*20000,
		),
	)
}
