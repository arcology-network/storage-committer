package commutative

import (
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
)

func (this *Int64) HeaderSize() uint32 {
	return 0 //static size only , no header needed,
}

func (this *Int64) Size() uint32 {
	return this.HeaderSize() + // No need to encode this.finalized
		codec.Bool(this.finalized).Size() +
		codec.Int64(this.value).Size() +
		codec.Int64(this.delta).Size()
}

func (this *Int64) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Int64) EncodeToBuffer(buffer []byte) int {
	offset := codec.Bool(this.finalized).EncodeToBuffer(buffer)
	offset += codec.Int64(this.value).EncodeToBuffer(buffer[offset:])
	return offset + codec.Int64(this.delta).EncodeToBuffer(buffer[offset:])
}

func (this *Int64) Decode(buffer []byte) interface{} {
	this = NewInt64(int64(
		codec.Int64(0).Decode(buffer[1:codec.UINT64_LEN+1]).(codec.Int64)),
		int64(codec.Int64(0).Decode(buffer[codec.UINT64_LEN+1:]).(codec.Int64))).(*Int64)

	this.finalized = bool(codec.Bool(this.finalized).Decode(buffer).(codec.Bool))
	return this
}

func (this *Int64) EncodeCompact() []byte {
	return codec.Int64(this.value).Encode()
}

func (this *Int64) DecodeCompact(bytes []byte) interface{} {
	value := NewInt64(0, 0)
	value.(*Int64).value = int64(codec.Int64(0).Decode(bytes).(codec.Int64))
	return value
}

func (this *Int64) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}

func (this *Int64) GetDelta() interface{} {
	return this.delta
}
