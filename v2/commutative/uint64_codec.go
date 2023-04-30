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

func (this *Uint64) Size(selector ...bool) uint32 {
	return this.HeaderSize() +
		common.IfThen(len(selector) == 0 || selector[0], uint32(8), 0) +
		common.IfThen(len(selector) == 0 || selector[1], uint32(8), 0) +
		common.IfThen(len(selector) == 0 || selector[2], uint32(8), 0) +
		common.IfThen(len(selector) == 0 || selector[3], uint32(8), 0)
}

func (this *Uint64) Encode(selector ...bool) []byte {
	buffer := make([]byte, this.Size(selector...))
	offset := codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			common.IfThen(len(selector) == 0 || selector[0], uint32(8), 0),
			common.IfThen(len(selector) == 0 || selector[1], uint32(8), 0),
			common.IfThen(len(selector) == 0 || selector[2], uint32(8), 0),
			common.IfThen(len(selector) == 0 || selector[3], uint32(8), 0),
		},
	)
	this.EncodeToBuffer(buffer[offset:], selector...)
	return buffer
}

func (this *Uint64) EncodeToBuffer(buffer []byte, selector ...bool) int {
	offset := common.IfThenDo1st(len(selector) == 0 || selector[0], func() int { return this.value.EncodeToBuffer(buffer) }, 0)
	offset += common.IfThenDo1st(len(selector) == 0 || selector[1], func() int { return this.delta.EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(len(selector) == 0 || selector[2], func() int { return this.min.EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(len(selector) == 0 || selector[3], func() int { return this.max.EncodeToBuffer(buffer[offset:]) }, 0)
	return offset
}

func (this *Uint64) Decode(buffer []byte) interface{} {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)

	this.value = codec.Uint64(0).Decode(fields[0]).(codec.Uint64)
	this.delta = codec.Uint64(0).Decode(fields[1]).(codec.Uint64)
	this.min = codec.Uint64(0).Decode(fields[2]).(codec.Uint64)
	this.max = codec.Uint64(math.MaxUint64).Decode(fields[3]).(codec.Uint64)
	return this
}

func (this *Uint64) EncodeCompact() []byte {
	return this.Encode()
}

func (this *Uint64) DecodeCompact(buffer []byte) interface{} {
	return (&Uint64{}).Decode(buffer)
}

func (this *Uint64) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}
