package commutative

import (
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"

	// performance "github.com/arcology-network/common-lib/mhasher"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	orderedmap "github.com/elliotchance/orderedmap"
)

func (this *Meta) Encode() []byte {
	this.keys = this.keys[:0] // Clear keys, no need to send

	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Meta) HeaderSize() uint32 {
	return 5 * codec.UINT32_LEN
}

func (this *Meta) Size() uint32 {
	if this == nil {
		return 0
	}

	total := this.HeaderSize() +
		codec.Strings(this.keys).Size() +
		codec.Strings(this.added).Size() +
		codec.Strings(this.removed).Size() +
		uint32(codec.Bool(this.finalized).Size())
	return total
}

func (this *Meta) FillHeader(buffer []byte) {
	total := uint32(0)
	codec.Uint32(4).EncodeToBuffer(buffer[codec.UINT32_LEN*0:])

	codec.Uint32(total).EncodeToBuffer(buffer[codec.UINT32_LEN*1:])
	total += codec.Strings(this.keys).Size()

	codec.Uint32(total).EncodeToBuffer(buffer[codec.UINT32_LEN*2:])
	total += codec.Strings(this.added).Size()

	codec.Uint32(total).EncodeToBuffer(buffer[codec.UINT32_LEN*3:])
	total += codec.Strings(this.removed).Size()

	codec.Uint32(total).EncodeToBuffer(buffer[codec.UINT32_LEN*4:])
}

func (this *Meta) EncodeToBuffer(buffer []byte) int {
	this.FillHeader(buffer)
	offset := int(this.HeaderSize())

	offset += codec.Strings(this.keys).EncodeToBuffer(buffer[offset:])
	offset += codec.Strings(this.added).EncodeToBuffer(buffer[offset:])
	offset += codec.Strings(this.removed).EncodeToBuffer(buffer[offset:])

	return int(offset) + codec.Bool(this.finalized).EncodeToBuffer(buffer[offset:])
}

func (this *Meta) Decode(bytes []byte) interface{} {
	buffers := codec.Byteset{}.Decode(bytes).(codec.Byteset)
	this = &Meta{
		keys:          codec.Strings([]string{}).Decode(common.ArrayCopy(buffers[0])).(codec.Strings),
		added:         codec.Strings([]string{}).Decode(common.ArrayCopy(buffers[1])).(codec.Strings),
		removed:       codec.Strings([]string{}).Decode(buffers[2]).(codec.Strings),
		finalized:     bool(codec.Bool(true).Decode(buffers[3]).(codec.Bool)),
		keyView:       nil,
		addedBuffer:   orderedmap.NewOrderedMap(),
		removedBuffer: make(map[string]ccurlcommon.UnivalueInterface),
		cacheDirty:    false,
	}
	return this
}

func (this *Meta) EncodeCompact() []byte {
	byteset := [][]byte{
		codec.Strings(this.keys).Encode(),
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Meta) DecodeCompact(bytes []byte) interface{} {
	buffers := codec.Byteset{}.Decode(bytes).(codec.Byteset)
	return &Meta{
		keys:          codec.Strings([]string{}).Decode(buffers[0]).(codec.Strings),
		added:         []string{},
		removed:       []string{},
		finalized:     false,
		keyView:       nil,
		addedBuffer:   orderedmap.NewOrderedMap(),
		removedBuffer: make(map[string]ccurlcommon.UnivalueInterface),
		cacheDirty:    false,
	}
}

func (this *Meta) Print() {
	fmt.Println("Keys: ", this.keys)
	fmt.Println("Added: ", this.added)
	fmt.Println("Removed: ", this.removed)
	fmt.Println()
}
