package commutative

import (
	"fmt"
	"math"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
)

func (this *Uint64) HeaderSize() uint32 {
	return 5 * codec.UINT32_LEN //static size only , no header needed,
}

func (this *Uint64) Size() uint32 {
	return this.HeaderSize() +
		common.IfThen(this.value != nil, uint32(8), 0) +
		common.IfThen(this.delta != nil, uint32(8), 0) +
		common.IfThen(this.min != nil, uint32(8), 0) +
		common.IfThen(this.max != nil, uint32(8), 0)
}

func (this *Uint64) Encode() []byte {
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

func (this *Uint64) EncodeToBuffer(buffer []byte) int {
	offset := common.IfThenDo1st(this.value != nil, func() int { return this.value.EncodeToBuffer(buffer) }, 0)
	offset += common.IfThenDo1st(this.delta != nil, func() int { return this.delta.EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(this.min != nil, func() int { return this.min.EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(this.max != nil, func() int { return this.max.EncodeToBuffer(buffer[offset:]) }, 0)
	return offset
}

func (this *Uint64) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return this
	}

	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)

	value := codec.Uint64(0).Decode(fields[0]).(codec.Uint64)
	delta := codec.Uint64(0).Decode(fields[1]).(codec.Uint64)
	min := codec.Uint64(0).Decode(fields[2]).(codec.Uint64)
	max := codec.Uint64(math.MaxUint64).Decode(fields[3]).(codec.Uint64)

	this.value = &value
	this.delta = &delta
	this.min = &min
	this.max = &max
	return this
}

func (this *Uint64) Print() {
	fmt.Println(" Value: ", *this.value, "Delta: ", *this.delta)
}
