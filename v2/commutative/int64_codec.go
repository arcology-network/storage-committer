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

func (this *Int64) Size() uint32 {
	return this.HeaderSize() +
		common.IfThen(this.value != nil, uint32(8), 0) +
		common.IfThen(this.delta != nil, uint32(8), 0) +
		common.IfThen(this.min != nil, uint32(8), 0) +
		common.IfThen(this.max != nil, uint32(8), 0)
}

func (this *Int64) Encode() []byte {
	buffer := make([]byte, this.Size())
	offset := codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			common.IfThen(this.value != nil, uint32(8), 0),
			common.IfThen(this.delta != nil, uint32(8), 0),
			common.IfThen(this.min != nil, uint32(8), 0),
			common.IfThen(this.max != nil, uint32(8), 0),
		},
	)
	this.EncodeToBuffer(buffer[offset:])
	return buffer
}

func (this *Int64) EncodeToBuffer(buffer []byte) int {
	offset := common.IfThenDo1st(this.value != nil, func() int { return this.value.EncodeToBuffer(buffer) }, 0)
	offset += common.IfThenDo1st(this.delta != nil, func() int { return this.delta.EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(this.min != nil, func() int { return this.min.EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(this.max != nil, func() int { return this.max.EncodeToBuffer(buffer[offset:]) }, 0)
	return offset
}

func (this *Int64) Decode(buffer []byte) interface{} {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)

	value := codec.Int64(0).Decode(fields[0]).(codec.Int64)
	delta := codec.Int64(0).Decode(fields[1]).(codec.Int64)
	min := codec.Int64(math.MinInt64).Decode(fields[2]).(codec.Int64)
	max := codec.Int64(math.MaxInt64).Decode(fields[3]).(codec.Int64)

	this.value = &value
	this.delta = &delta
	this.min = &min
	this.max = &max
	return this
}

func (this *Int64) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}
