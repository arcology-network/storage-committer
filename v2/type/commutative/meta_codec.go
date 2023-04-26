package commutative

import (
	"bytes"
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"

	// performance "github.com/arcology-network/common-lib/mhasher"
	orderedset "github.com/arcology-network/common-lib/container/set"
)

func (this *Meta) Encode(processors ...func(interface{}) interface{}) []byte {
	buffer := make([]byte, this.Size()) //  no need to send the committed keys
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Meta) HeaderSize() uint32 {
	return 3 * codec.UINT32_LEN // number of fields + 1
}

func (this *Meta) Size() uint32 {
	if this == nil {
		return 0
	}

	total := this.HeaderSize() +
		codec.Strings(this.addDict.Keys()).Size() +
		codec.Strings(this.delDict.Keys()).Size()
	return total
}

func (this *Meta) FillHeader(buffer []byte) {
	codec.Uint32(2).EncodeToBuffer(buffer) // number of fields

	offset := uint32(0)
	codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*1:])
	offset += codec.Strings(this.addDict.Keys()).Size()

	codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*2:])
}

func (this *Meta) EncodeToBuffer(buffer []byte, processors ...func(interface{}) interface{}) int {
	this.FillHeader(buffer)
	offset := int(this.HeaderSize())

	offset += codec.Strings(this.addDict.Keys()).EncodeToBuffer(buffer[offset:])
	offset += codec.Strings(this.delDict.Keys()).EncodeToBuffer(buffer[offset:])

	return int(offset)
}

func (this *Meta) Decode(buffer []byte) interface{} {
	buffers := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	this = &Meta{
		value:   orderedset.NewOrderedSet([]string{}),
		addDict: orderedset.NewOrderedSet(codec.Strings([]string{}).Decode(bytes.Clone(buffers[0])).(codec.Strings)),
		delDict: orderedset.NewOrderedSet(codec.Strings([]string{}).Decode(bytes.Clone(buffers[1])).(codec.Strings)),
	}

	return this
}

func (this *Meta) EncodeCompact() []byte {
	byteset := [][]byte{
		codec.Strings(this.value.Keys()).Encode(),
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Meta) DecodeCompact(bytes []byte) interface{} {
	buffers := codec.Byteset{}.Decode(bytes).(codec.Byteset)
	return &Meta{
		value:   orderedset.NewOrderedSet(codec.Strings([]string{}).Decode(buffers[0]).(codec.Strings)),
		addDict: orderedset.NewOrderedSet([]string{}),
		delDict: orderedset.NewOrderedSet([]string{}),
	}
}

func (this *Meta) Print() {
	// fmt.Println("Keys: ", this.committedKeys)
	fmt.Println("Added: ", this.addDict.Keys())
	fmt.Println("Removed: ", this.delDict.Keys())
	fmt.Println()
}
