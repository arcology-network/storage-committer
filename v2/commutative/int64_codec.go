package commutative

import (
	"fmt"
	"math"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
)

func (this *Int64) HeaderSize() uint32 {
	return 5 * codec.UINT32_LEN //static size only , no header needed,
}

func (this *Int64) Size(selectors ...interface{}) uint32 {
	return this.HeaderSize() +
		common.IfThen(len(selectors) == 0 || selectors[0].(bool), uint32(8), 0) +
		common.IfThen(len(selectors) == 0 || selectors[1].(bool), uint32(8), 0) +
		common.IfThen(len(selectors) == 0 || selectors[2].(bool), uint32(8), 0) +
		common.IfThen(len(selectors) == 0 || selectors[3].(bool), uint32(8), 0)
}

func (this *Int64) Encode(selectors ...interface{}) []byte {
	buffer := make([]byte, this.Size(selectors...))
	offset := codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			common.IfThen(len(selectors) == 0 || selectors[0].(bool), uint32(8), 0),
			common.IfThen(len(selectors) == 0 || selectors[1].(bool), uint32(8), 0),
			common.IfThen(len(selectors) == 0 || selectors[2].(bool), uint32(8), 0),
			common.IfThen(len(selectors) == 0 || selectors[3].(bool), uint32(8), 0),
		},
	)
	this.EncodeToBuffer(buffer[offset:], selectors...)
	return buffer
}

func (this *Int64) EncodeToBuffer(buffer []byte, selectors ...interface{}) int {
	offset := common.IfThenDo1st(len(selectors) == 0 || selectors[0].(bool), func() int { return this.value.EncodeToBuffer(buffer) }, 0)
	offset += common.IfThenDo1st(len(selectors) == 0 || selectors[1].(bool), func() int { return this.delta.EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(len(selectors) == 0 || selectors[2].(bool), func() int { return this.min.EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(len(selectors) == 0 || selectors[3].(bool), func() int { return this.max.EncodeToBuffer(buffer[offset:]) }, 0)
	return offset
}

func (this *Int64) Decode(buffer []byte) interface{} {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)

	this.value = codec.Int64(0).Decode(fields[0]).(codec.Int64)
	this.delta = codec.Int64(0).Decode(fields[1]).(codec.Int64)
	this.min = codec.Int64(math.MinInt64).Decode(fields[2]).(codec.Int64)
	this.max = codec.Int64(math.MaxInt64).Decode(fields[3]).(codec.Int64)
	return this
}

// func (this *Int64) Encode() []byte {
// 	return this.Encode()
// }

// func (this *Int64) DecodeCompact(buffer []byte) interface{} {
// 	return (&Uint64{}).Decode(buffer)
// }

func (this *Int64) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}
