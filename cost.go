package concurrenturl

import (
	common "github.com/arcology-network/common-lib/common"
	// ethparams "github.com/arcology-network/evm/params"
)

const (
	IS_PATH = uint64(3)
)

type Cost struct{}

func (Cost) Reader(dataSize uint64, hotLoad bool) uint64 { // Call this before setting the value attribute to nil
	return common.IfThen(
		hotLoad,
		common.Max(dataSize/32, 1)*3,
		(dataSize/32)*5000,
	)
}

func (Cost) Writer(dataSize, committedSize uint64) int64 { // May get refunds sometimes
	return common.IfThen(
		committedSize > 0,
		common.Max(int64(dataSize/32), 1)*50000,
		common.IfThen(
			int64(dataSize) > int64(committedSize),
			int64(dataSize),
			((int64(committedSize)-int64(dataSize))/32)*50000,
		),
	)
}
