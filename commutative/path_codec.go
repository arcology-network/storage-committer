package commutative

import (
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/ethereum/go-ethereum/rlp"

	orderedset "github.com/arcology-network/common-lib/container/set"
)

func (this *Path) HeaderSize() uint32 {
	return 3 * codec.UINT32_LEN // number of fields + 1
}

func (this *Path) Size() uint32 {
	return this.HeaderSize() +
		common.IfThenDo1st(this.value != nil, func() uint32 { return this.value.Size() }, 0) +
		common.IfThenDo1st(this.delta != nil, func() uint32 { return this.delta.Size() }, 0)
}

func (this *Path) Encode() []byte {
	buffer := make([]byte, this.Size()) //  no need to send the committed keys
	offset := codec.Encoder{}.FillHeader(buffer,
		[]uint32{
			common.IfThenDo1st(this.value != nil, func() uint32 { return this.value.Size() }, 0),
			common.IfThenDo1st(this.delta != nil, func() uint32 { return this.delta.Size() }, 0),
		},
	)
	this.EncodeToBuffer(buffer[offset:])
	return buffer
}

func (this *Path) EncodeToBuffer(buffer []byte) int {
	offset := common.IfThenDo1st(this.value != nil, func() int { return this.value.EncodeToBuffer(buffer) }, 0)
	offset += common.IfThenDo1st(this.delta != nil, func() int { return this.delta.EncodeToBuffer(buffer[offset:]) }, 0)
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

func (this *Path) StorageEncode() []byte {
	buffer, _ := rlp.EncodeToBytes(this.Encode())
	return buffer
}

func (this *Path) StorageDecode(buffer []byte) interface{} {
	var decoded []byte
	rlp.DecodeBytes(buffer, &decoded)
	return this.Decode(decoded)
}
