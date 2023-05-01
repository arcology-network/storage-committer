package commutative

import (
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
	orderedset "github.com/arcology-network/common-lib/container/set"
)

func (this *PathDelta) Encode(...interface{}) []byte {
	buffer := make([]byte, this.Size()) //  no need to send the committed keys
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *PathDelta) Size(...interface{}) uint32 {
	return codec.Stringset([][]string{this.addDict.Keys(), this.delDict.Keys()}).Size()
}

func (this *PathDelta) EncodeToBuffer(buffer []byte, _ ...interface{}) int {
	return codec.Stringset([][]string{this.addDict.Keys(), this.delDict.Keys()}).EncodeToBuffer(buffer)
}

func (this *PathDelta) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return this
	}

	fields := codec.Stringset{}.Decode(buffer).(codec.Stringset)
	this = &PathDelta{
		addDict: orderedset.NewOrderedSet(fields[0]),
		delDict: orderedset.NewOrderedSet(fields[1]),
	}
	return this
}

func (this *PathDelta) Print() {
	// fmt.Println("Keys: ", this.committedKeys)
	fmt.Println("Added: ", this.addDict.Keys())
	fmt.Println("Removed: ", this.delDict.Keys())
	fmt.Println()
}
