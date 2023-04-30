package commutative

import (
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"

	// performance "github.com/arcology-network/common-lib/mhasher"
	orderedset "github.com/arcology-network/common-lib/container/set"
)

func (this *Path) HeaderSize() uint32 {
	return 3 * codec.UINT32_LEN // number of fields + 1
}

func (this *Path) Size(selectors ...bool) uint32 {
	return common.IfThen(len(selectors) == 0 || selectors[0], this.value.Size(), 0) +
		common.IfThen(len(selectors) == 0 || selectors[1], this.delta.Size(), 0)
}

func (this *Path) Encode(selectors ...bool) []byte {
	buffer := make([]byte, this.Size()) //  no need to send the committed keys
	offset := codec.Encoder{}.FillHeader(buffer,
		[]uint32{
			common.IfThen(len(selectors) == 0 || selectors[0], this.value.Size(), 0),
			common.IfThen(len(selectors) == 0 || selectors[0], this.delta.Size(), 0),
		},
	)
	this.EncodeToBuffer(buffer[offset:], selectors...)
	return buffer
}

func (this *Path) EncodeToBuffer(buffer []byte, selectors ...bool) int {
	offset := common.IfThen(len(selectors) == 0 || selectors[0], this.value.EncodeToBuffer(buffer), 0)
	offset += common.IfThen(len(selectors) == 0 || selectors[1], this.delta.EncodeToBuffer(buffer[offset:]), 0)
	return offset
}

func (this *Path) Decode(buffer []byte) interface{} {
	return &Path{
		value: orderedset.NewOrderedSet([]string{}),
		delta: (&PathDelta{}).Decode(buffer).(*PathDelta),
	}
}

func (this *Path) Print() {
	// fmt.Println("Keys: ", this.committedKeys)
	fmt.Println("Added: ", this.delta.addDict.Keys())
	fmt.Println("Removed: ", this.delta.delDict.Keys())
	fmt.Println()
}
