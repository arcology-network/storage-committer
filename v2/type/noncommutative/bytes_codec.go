package noncommutative

import (
	"bytes"
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
)

func (this *Bytes) HeaderSize() uint32 {
	return 3 * codec.UINT32_LEN
}

func (this *Bytes) Size() uint32 {
	return this.HeaderSize() + uint32(1+len(this.data))
}

func (this *Bytes) Encode(processors ...func(interface{}) interface{}) []byte {
	byteset := [][]byte{
		codec.Bool(this.placeholder).Encode(),
		this.data,
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Bytes) EncodeToBuffer(buffer []byte, processors ...func(interface{}) interface{}) int {
	offset := codec.Bool(this.placeholder).EncodeToBuffer(buffer)
	return offset + codec.Bytes(this.data).EncodeToBuffer(buffer[offset:])
}

func (*Bytes) Decode(buffer []byte) interface{} {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	return &Bytes{
		placeholder: bool(codec.Bool(true).Decode(fields[0]).(codec.Bool)),
		data:        bytes.Clone(fields[1]),
	}
}

func (this *Bytes) EncodeCompact() []byte {
	return this.Encode()
}

func (this *Bytes) DecodeCompact(bytes []byte) interface{} {
	return this.Decode(bytes)
}

func (this *Bytes) Purge() {}

func (this *Bytes) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

func (this *Bytes) Print() {
	fmt.Println(*this)
	fmt.Println()
}
