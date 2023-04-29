package commutative

import (
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
)

func (this *Int64) HeaderSize() uint32 {
	return 0 //static size only , no header needed,
}

func (this *Int64) Size() uint32 {
	return codec.Int64(this.value).Size() +
		codec.Int64(this.delta).Size() +
		codec.Int64(this.min).Size() +
		codec.Int64(this.max).Size()
}

func (this *Int64) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Int64) EncodeToBuffer(buffer []byte) int {
	offset := codec.Int64(this.value).EncodeToBuffer(buffer)
	offset += codec.Int64(this.delta).EncodeToBuffer(buffer[offset:])
	offset += codec.Int64(this.min).EncodeToBuffer(buffer[offset:])
	offset += codec.Int64(this.max).EncodeToBuffer(buffer[offset:])
	return offset
}

func (this *Int64) Decode(buffer []byte) interface{} {
	this = &Int64{
		int64(codec.Int64(0).Decode(buffer[:]).(codec.Int64)),                  // value
		int64(codec.Int64(0).Decode(buffer[codec.INT64_LEN*1:]).(codec.Int64)), // delta
		int64(codec.Int64(0).Decode(buffer[codec.INT64_LEN*2:]).(codec.Int64)), // min
		int64(codec.Int64(0).Decode(buffer[codec.INT64_LEN*3:]).(codec.Int64)), // max

	}
	return this
}

func (this *Int64) EncodeCompact() []byte {
	return this.Encode()
}

func (this *Int64) DecodeCompact(buffer []byte) interface{} {
	return (&Int64{}).Decode(buffer)
}

func (this *Int64) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}
