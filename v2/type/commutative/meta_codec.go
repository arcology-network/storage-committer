package commutative

import (
	"bytes"
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"

	// performance "github.com/arcology-network/common-lib/mhasher"

	orderedmap "github.com/elliotchance/orderedmap"
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
		codec.Strings(this.added).Size() +
		codec.Strings(this.removed).Size()
	return total
}

func (this *Meta) FillHeader(buffer []byte) {
	codec.Uint32(2).EncodeToBuffer(buffer) // number of fields

	offset := uint32(0)
	codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*1:])
	offset += codec.Strings(this.added).Size()

	codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*2:])
	//offset += codec.Strings(this.removed).Size()
}

func (this *Meta) EncodeToBuffer(buffer []byte, processors ...func(interface{}) interface{}) int {
	this.FillHeader(buffer)
	offset := int(this.HeaderSize())

	offset += codec.Strings(this.added).EncodeToBuffer(buffer[offset:])
	offset += codec.Strings(this.removed).EncodeToBuffer(buffer[offset:])

	return int(offset)
}

func (this *Meta) Decode(buffer []byte) interface{} {
	buffers := codec.Byteset{}.Decode(buffer).(codec.Byteset)

	this = &Meta{
		committedKeys: []string{},
		added:         codec.Strings([]string{}).Decode(bytes.Clone(buffers[0])).(codec.Strings),
		removed:       codec.Strings([]string{}).Decode(buffers[1]).(codec.Strings),
		view:          nil,
		addedBuffer:   orderedmap.NewOrderedMap(),
		removedBuffer: orderedmap.NewOrderedMap(),
		snapshotDirty: false,
	}

	return this
}

func (this *Meta) EncodeCompact() []byte {
	byteset := [][]byte{
		codec.Strings(this.committedKeys).Encode(),
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Meta) DecodeCompact(bytes []byte) interface{} {
	buffers := codec.Byteset{}.Decode(bytes).(codec.Byteset)
	return &Meta{
		committedKeys: codec.Strings([]string{}).Decode(buffers[0]).(codec.Strings),
		added:         []string{},
		removed:       []string{},
		view:          nil,
		addedBuffer:   orderedmap.NewOrderedMap(),
		removedBuffer: orderedmap.NewOrderedMap(),
		snapshotDirty: false,
	}
}

func (this *Meta) Print() {
	fmt.Println("Keys: ", this.committedKeys)
	fmt.Println("Added: ", this.added)
	fmt.Println("Removed: ", this.removed)
	fmt.Println()
}
