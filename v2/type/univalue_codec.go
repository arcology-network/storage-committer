package ccurltype

import (
	"reflect"

	codec "github.com/HPISTechnologies/common-lib/codec"
	common "github.com/HPISTechnologies/common-lib/common"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
)

func (this *Univalue) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *Univalue) HeaderSize() uint32 {
	return uint32(10 * codec.UINT32_LEN)
}

func (this *Univalue) Size() uint32 {
	vLen := uint32(0)
	if this.value != nil {
		vLen = (this.value.(ccurlcommon.TypeInterface).Size())
	}

	return this.HeaderSize() + // uint32(9*codec.UINT32_LEN) +
		uint32(1) + // transition type id
		uint32(1) + // codec.Uint8(this.vType).Size() +
		uint32(4) + // codec.Uint32(uint32(this.tx)).Size() +
		uint32(len(*this.path)) + // codec.String(*this.path).Size() +
		uint32(4) + // codec.Uint32(this.reads).Size() +
		uint32(4) + // codec.Uint32(this.writes).Size() +
		(vLen) +
		uint32(1) + // codec.Bool(this.preexists).Size() +
		uint32(1) // codec.Bool(this.composite).Size()
}

func (this *Univalue) FillHeader(buffer []byte) int {
	vLen := uint32(0)
	if this.value != nil {
		vLen = this.value.(ccurlcommon.TypeInterface).Size()
		if uint32(len(this.value.(ccurlcommon.TypeInterface).Encode())) != this.value.(ccurlcommon.TypeInterface).Size() {
			panic("Error: Sizes don't match")
		}
	}

	return codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			uint32(codec.Uint8(this.transitType).Size()),
			uint32(codec.Uint8(this.vType).Size()),
			codec.Uint32(this.tx).Size(),
			codec.String(*this.path).Size(),
			codec.Uint32(this.reads).Size(),
			codec.Uint32(this.writes).Size(),
			vLen,
			codec.Bool(this.preexists).Size(),
			codec.Bool(this.composite).Size(),
		},
	)
}

func (this *Univalue) EncodeToBuffer(buffer []byte) int {
	offset := this.FillHeader(buffer)
	offset += codec.Uint8(this.transitType).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint8(this.vType).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.tx).EncodeToBuffer(buffer[offset:])
	offset += codec.String(*this.path).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.reads).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.writes).EncodeToBuffer(buffer[offset:])
	if this.value != nil {
		offset += codec.Bytes(this.value.(ccurlcommon.TypeInterface).Encode()).EncodeToBuffer(buffer[offset:])
	}

	offset += codec.Bool(this.preexists).EncodeToBuffer(buffer[offset:])
	offset += codec.Bool(this.composite).EncodeToBuffer(buffer[offset:])
	return offset
}

func (this *Univalue) Decode(buffer []byte) interface{} {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	if len(fields) == 1 {
		return this
	}

	this.transitType = uint8(reflect.Kind(codec.Uint8(0).Decode(fields[0]).(codec.Uint8)))
	this.vType = uint8(reflect.Kind(codec.Uint8(1).Decode(fields[1]).(codec.Uint8)))
	this.tx = uint32(codec.Uint32(1).Decode(fields[2]).(codec.Uint32))
	key := string(codec.String("").Decode(common.ArrayCopy(fields[3])).(codec.String))
	this.path = &key
	this.reads = uint32(codec.Uint32(1).Decode(fields[4]).(codec.Uint32))
	this.writes = uint32(codec.Uint32(1).Decode(fields[5]).(codec.Uint32))
	this.value = (&Decoder{}).Decode(fields[6], this.vType)
	this.preexists = bool(codec.Bool(true).Decode(fields[7]).(codec.Bool))
	this.composite = bool(codec.Bool(true).Decode(fields[8]).(codec.Bool))
	this.reserved = fields[6] // For merkle root calculation

	if this.value == nil || this.IsCommutative() {
		this.reserved = nil
	}
	return this
}

func (this *Univalue) GetEncoded() []byte {
	if this.value == nil {
		return []byte{}
	}

	if this.IsCommutative() {
		return this.value.(ccurlcommon.TypeInterface).EncodeCompact()
	}

	if this.reserved == nil {
		return this.value.(ccurlcommon.TypeInterface).EncodeCompact()
	}
	return this.reserved.([]byte)
}

func (this *Univalue) GetEncodedSize() []int {
	vBytes := []byte{}
	if this.value != nil {
		vBytes = this.value.(ccurlcommon.TypeInterface).Encode()
	}

	return []int{
		int(codec.Uint8(this.transitType).Size()),
		int(codec.Uint8(this.vType).Size()),
		int(codec.Uint32(this.tx).Size()),
		int(codec.String(*this.path).Size()),
		int(codec.Uint32(this.reads).Size()),
		int(codec.Uint32(this.writes).Size()),
		len(vBytes),
		int(codec.Bool(this.preexists).Size()),
		int(codec.Bool(this.composite).Size()),
	}
}

func (this *Univalue) GobEncode() ([]byte, error) {
	return this.Encode(), nil
}

func (this *Univalue) GobDecode(data []byte) error {
	v := this.Decode(data).(*Univalue)
	this.vType = v.vType
	this.composite = v.composite
	this.path = v.path
	this.preexists = v.preexists
	this.reads = v.reads
	this.tx = v.tx
	this.value = v.value
	this.writes = v.writes
	return nil
}
