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

func (this *Path) Size(selectors ...interface{}) uint32 {
	return this.HeaderSize() +
		common.IfThenDo1st(len(selectors) == 0 || selectors[0].(bool), func() uint32 { return this.value.Size() }, 0) +
		common.IfThenDo1st(len(selectors) == 0 || selectors[1].(bool), func() uint32 { return this.delta.Size() }, 0)
}

func (this *Path) Encode(selectors ...interface{}) []byte {
	buffer := make([]byte, this.Size(selectors...)) //  no need to send the committed keys
	offset := codec.Encoder{}.FillHeader(buffer,
		[]uint32{
			common.IfThen(len(selectors) == 0 || selectors[0].(bool), this.value.Size(), 0),
			common.IfThen(len(selectors) == 0 || selectors[0].(bool), this.delta.Size(), 0),
		},
	)
	this.EncodeToBuffer(buffer[offset:], selectors...)
	return buffer
}

func (this *Path) EncodeToBuffer(buffer []byte, selectors ...interface{}) int {
	offset := common.IfThen(len(selectors) == 0 || selectors[0].(bool), this.value.EncodeToBuffer(buffer), 0)
	offset += common.IfThenDo1st(len(selectors) == 0 || selectors[1].(bool), func() int { return this.delta.EncodeToBuffer(buffer[offset:]) }, 0)
	return offset
}

func (this *Path) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return this
	}

	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	return &Path{
		value: orderedset.NewOrderedSet(codec.Strings{}.Decode(fields[0]).(codec.Strings)),
		delta: NewPathDelta([]string{}, []string{}).Decode(fields[1]).(*PathDelta),
	}
}

func (this *Path) Print() {
	// fmt.Println("Keys: ", this.committedKeys)
	fmt.Println("Added: ", this.delta.addDict.Keys())
	fmt.Println("Removed: ", this.delta.delDict.Keys())
	fmt.Println()
}
