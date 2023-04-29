package commutative

import (
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
)

func (this *Uint64) HeaderSize() uint32 {
	return 0 //static size only , no header needed,
}

func (this *Uint64) Size() uint32 {
	return this.HeaderSize() + // No need to encode this.finalized
		codec.Uint64(this.value).Size() +
		codec.Uint64(this.delta).Size() +
		codec.Uint64(this.min).Size() +
		codec.Uint64(this.max).Size()
}

func (this *Uint64) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Uint64) EncodeToBuffer(buffer []byte) int {
	offset := 0
	offset += codec.Uint64(this.value).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint64(this.delta).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint64(this.min).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint64(this.max).EncodeToBuffer(buffer[offset:])
	return offset
}

func (this *Uint64) Decode(buffer []byte) interface{} {
	this = &Uint64{
		uint64(codec.Uint64(0).Decode(buffer).(codec.Uint64)),                      // value
		uint64(codec.Uint64(0).Decode(buffer[codec.UINT64_LEN*1:]).(codec.Uint64)), // delta
		uint64(codec.Uint64(0).Decode(buffer[codec.UINT64_LEN*2:]).(codec.Uint64)), // min
		uint64(codec.Uint64(0).Decode(buffer[codec.UINT64_LEN*3:]).(codec.Uint64)), // max
	}
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
