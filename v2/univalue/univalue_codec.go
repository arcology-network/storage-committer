package univalue

import (
	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

func (this *Univalue) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Univalue) HeaderSize() uint32 {
	return uint32(4 * codec.UINT32_LEN)
}

func (this *Univalue) Sizes() []int {
	return []int{
		int(this.HeaderSize()),
		int(this.Unimeta.Size()),
		int(this.value.(ccurlcommon.TypeInterface).Size()),
		int(codec.Bytes(this.cache).Size()),
	}
}

func (this *Univalue) Size() uint32 {
	return this.HeaderSize() + this.Unimeta.Size() +
		common.IfThenDo1st(this.value != nil, func() uint32 { return this.value.(ccurlcommon.TypeInterface).Size() }, 0) +
		codec.Bytes(this.cache).Size()
}

func (this *Univalue) FillHeader(buffer []byte) int {
	return codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			this.Unimeta.Size(),
			common.IfThenDo1st(this.value != nil, func() uint32 { return this.value.(ccurlcommon.TypeInterface).Size() }, 0),
			codec.Bytes(this.cache).Size(),
		},
	)
}

func (this *Univalue) EncodeToBuffer(buffer []byte) int {
	offset := this.FillHeader(buffer)

	offset += this.Unimeta.EncodeToBuffer(buffer[offset:])
	offset += common.IfThenDo1st(this.value != nil, func() int {
		return codec.Bytes(this.value.(ccurlcommon.TypeInterface).Encode()).EncodeToBuffer(buffer[offset:])
	}, 0)
	offset += codec.Bytes{}.EncodeToBuffer(buffer[offset:])

	return offset
}

func (this *Univalue) Decode(buffer []byte) interface{} {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	unimeta := (&Unimeta{}).Decode(fields[0]).(*Unimeta)

	return &Univalue{
		*unimeta,
		common.IfThenDo1st(len(fields) > 1, func() interface{} {
			return (&Decoder{}).Decode(fields[1], unimeta.vType)
		}, nil),
		(&codec.Bytes{}).Decode(fields[2]).(codec.Bytes),
	}
}

func (this *Univalue) GetEncoded() []byte {
	if this.value == nil {
		return []byte{}
	}

	if this.IsCommutative(this) {
		return this.value.(ccurlcommon.TypeInterface).EncodeCompact()
	}

	if len(this.cache) > 0 {
		return this.value.(ccurlcommon.TypeInterface).EncodeCompact()
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
