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

func (this *Uint64) Size(selectors ...interface{}) uint32 {
	return this.HeaderSize() +
		common.IfThen(len(selectors) == 0 || selectors[0].(bool), uint32(8), 0) +
		common.IfThen(len(selectors) == 0 || selectors[1].(bool), uint32(8), 0) +
		common.IfThen(len(selectors) == 0 || selectors[2].(bool), uint32(8), 0) +
		common.IfThen(len(selectors) == 0 || selectors[3].(bool), uint32(8), 0)
}

func (this *Uint64) Encode(selectors ...interface{}) []byte {
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

func (this *Uint64) EncodeToBuffer(buffer []byte, selectors ...interface{}) int {
	offset := common.IfThenDo1st(len(selectors) == 0 || selectors[0].(bool), func() int { return this.value.EncodeToBuffer(buffer) }, 0)
	offset += common.IfThenDo1st(len(selectors) == 0 || selectors[1].(bool), func() int { return this.delta.EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(len(selectors) == 0 || selectors[2].(bool), func() int { return this.min.EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(len(selectors) == 0 || selectors[3].(bool), func() int { return this.max.EncodeToBuffer(buffer[offset:]) }, 0)
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

// func (this *Uint64) Encode() []byte {
// 	return this.Encode()
// }

// func (this *Uint64) DecodeCompact(buffer []byte) interface{} {
// 	return (&Uint64{}).Decode(buffer)
// }

func (this *Uint64) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}
