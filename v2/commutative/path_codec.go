package commutative

import (
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"

	// performance "github.com/arcology-network/common-lib/mhasher"
	orderedset "github.com/arcology-network/common-lib/container/set"
)

func (this *Path) Encode() []byte {
	buffer := make([]byte, this.Size()) //  no need to send the committed keys
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Path) HeaderSize() uint32 {
	return 3 * codec.UINT32_LEN // number of fields + 1
}

func (this *Path) Size() uint32 {
	if this == nil {
		return 0
	}
	return this.delta.Size()
}

func (this *Path) EncodeToBuffer(buffer []byte) int {
	return int(this.delta.EncodeToBuffer(buffer))
}

func (this *Path) Decode(buffer []byte) interface{} {
	return &Path{
		value: orderedset.NewOrderedSet([]string{}),
		delta: (&PathDelta{}).Decode(buffer).(*PathDelta),
	}
}

func (this *Path) EncodeCompact() []byte {
	byteset := [][]byte{
		codec.Strings(this.value.Keys()).Encode(),
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Path) DecodeCompact(bytes []byte) interface{} {
	buffers := codec.Byteset{}.Decode(bytes).(codec.Byteset)
	return &Path{
		value: orderedset.NewOrderedSet(codec.Strings([]string{}).Decode(buffers[0]).(codec.Strings)),
		delta: NewPathDelta([]string{}, []string{}),
	}
}

func (this *Path) Print() {
	// fmt.Println("Keys: ", this.committedKeys)
	fmt.Println("Added: ", this.delta.addDict.Keys())
	fmt.Println("Removed: ", this.delta.delDict.Keys())
	fmt.Println()
}
