package univalue

import (
	"bytes"
	"reflect"

	codec "github.com/arcology-network/common-lib/codec"
)

func (this *Unimeta) Encode(processors ...interface{}) []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Unimeta) HeaderSize() uint32 {
	return uint32(8 * codec.UINT32_LEN)
}

func (this *Unimeta) Size() uint32 {
	return this.HeaderSize() + // uint32(9*codec.UINT32_LEN) +
		uint32(1) + // codec.Uint8(this.vType).Size() +
		uint32(4) + // codec.Uint32(uint32(this.tx)).Size() +
		uint32(len(*this.path)) + // codec.String(*this.path).Size() +
		uint32(4) + // codec.Uint32(this.reads).Size() +
		uint32(4) + // codec.Uint32(this.writes).Size() +
		uint32(4) + // codec.Uint32(this.deltaWrites).Size() +
		uint32(1) //+  codec.Bool(this.preexists).Size() +
}

func (this *Unimeta) FillHeader(buffer []byte) int {
	return codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			uint32(codec.Uint8(this.vType).Size()),
			codec.Uint32(this.tx).Size(),
			codec.String(*this.path).Size(),
			codec.Uint32(this.reads).Size(),
			codec.Uint32(this.writes).Size(),
			codec.Uint32(this.deltaWrites).Size(),
			codec.Bool(this.preexists).Size(),
		},
	)
}

func (this *Unimeta) EncodeToBuffer(buffer []byte, processors ...interface{}) int {
	offset := this.FillHeader(buffer)
	offset += codec.Uint8(this.vType).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.tx).EncodeToBuffer(buffer[offset:])
	offset += codec.String(*this.path).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.reads).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.writes).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.deltaWrites).EncodeToBuffer(buffer[offset:])
	offset += codec.Bool(this.preexists).EncodeToBuffer(buffer[offset:])

	return offset
}

func (this *Unimeta) Decode(buffer []byte) interface{} {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	if len(fields) == 1 {
		return this
	}

	this.vType = uint8(reflect.Kind(codec.Uint8(1).Decode(fields[0]).(codec.Uint8)))
	this.tx = uint32(codec.Uint32(0).Decode(fields[1]).(codec.Uint32))
	key := string(codec.String("").Decode(bytes.Clone(fields[2])).(codec.String))
	this.path = &key
	this.reads = uint32(codec.Uint32(1).Decode(fields[3]).(codec.Uint32))
	this.writes = uint32(codec.Uint32(1).Decode(fields[4]).(codec.Uint32))
	this.deltaWrites = uint32(codec.Uint32(1).Decode(fields[5]).(codec.Uint32))

	return this
}

// func (this *Unimeta) GetEncoded() []byte {
// 	if this.value == nil {
// 		return []byte{}
// 	}

// 	if this.IsCommutative() {
// 		return this.value.(ccurlcommon.TypeInterface).EncodeCompact()
// 	}

// 	if this.reserved == nil {
// 		return this.value.(ccurlcommon.TypeInterface).EncodeCompact()
// 	}
// 	return this.reserved.([]byte)
// }

// func (this *Unimeta) Sizes() []int {
// 	return []int{
// 		int(codec.Uint8(this.vType).Size()),
// 		int(codec.Uint32(this.tx).Size()),
// 		int(codec.String(*this.path).Size()),
// 		int(codec.Uint32(this.reads).Size()),
// 		int(codec.Uint32(this.writes).Size()),
// 		int(codec.Uint32(this.deltaWrites).Size()),
// 		int(codec.Bool(this.preexists).Size()),
// 	}
// }

func (this *Unimeta) GobEncode() ([]byte, error) {
	return this.Encode(), nil
}

func (this *Unimeta) GobDecode(data []byte) error {
	v := this.Decode(data).(*Unimeta)
	this.vType = v.vType
	this.path = v.path
	this.preexists = v.preexists
	this.tx = v.tx
	this.reads = v.reads
	this.writes = v.writes
	this.deltaWrites = v.deltaWrites
	return nil
}
