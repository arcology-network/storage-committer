package univalue

import (
	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"

	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
)

func (this *Univalue) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Univalue) HeaderSize() uint32 {
	return uint32(3 * codec.UINT32_LEN)
}

func (this *Univalue) Sizes() []uint32 {
	return []uint32{
		this.HeaderSize(),
		this.Property.Size(),
		this.value.(intf.Type).Size(),
	}
}

func (this *Univalue) Size() uint32 {
	return this.HeaderSize() +
		this.Property.Size() +
		common.IfThenDo1st(this.value != nil, func() uint32 { return this.value.(intf.Type).Size() }, 0)
}

func (this *Univalue) FillHeader(buffer []byte) int {
	return codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			this.Property.Size(),
			common.IfThenDo1st(this.value != nil, func() uint32 { return this.value.(intf.Type).Size() }, 0),
		},
	)
}

func (this *Univalue) EncodeToBuffer(buffer []byte) int {
	offset := this.FillHeader(buffer)

	offset += this.Property.EncodeToBuffer(buffer[offset:])
	offset += common.IfThenDo1st(this.value != nil, func() int {
		return codec.Bytes(this.value.(intf.Type).Encode()).EncodeToBuffer(buffer[offset:])
	}, 0)

	return offset
}

func (this *Univalue) Decode(buffer []byte) interface{} {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	property := (&Property{}).Decode(fields[0]).(*Property)

	return &Univalue{
		*property,
		(&platform.Codec{property.vType}).Decode(*property.path, fields[1], this.value),
		fields[1], // Keep copy, should expire as soon as the value is updated
	}
}

func (this *Univalue) GetEncoded() []byte {
	if this.value == nil {
		return []byte{}
	}

	if this.Value().(intf.Type).IsCommutative() {
		return this.value.(intf.Type).Value().(codec.Encodable).Encode()
	}

	if len(this.cache) > 0 {
		return this.value.(intf.Type).Value().(codec.Encodable).Encode()
	}
	return this.cache
}

func (this *Univalue) GobEncode() ([]byte, error) {
	return this.Encode(), nil
}

func (this *Univalue) GobDecode(buffer []byte) error {
	*this = *(&Univalue{}).Decode(buffer).(*Univalue)
	return nil
}
