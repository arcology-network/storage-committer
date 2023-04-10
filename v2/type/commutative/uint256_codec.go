package commutative

import (
	codec "github.com/arcology-network/common-lib/codec"
	uint256 "github.com/holiman/uint256"
)

func (this *U256) HeaderSize() uint32 {
	return (1 + 5) * codec.UINT32_LEN // Total number of fields + offsets of these fields
}

func (this *U256) Size() uint32 {
	return codec.Bool(this.finalized).Size() +
		32 + // Values
		32 + // Min
		32 + // Max
		32 + // delta
		1 // operation
}

func (this *U256) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *U256) EncodeToBuffer(buffer []byte) int {
	offset := codec.Bool(this.finalized).EncodeToBuffer(buffer)
	offset += codec.Uint64s(this.value[:]).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint64s(this.min[:]).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint64s(this.max[:]).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint64s(this.delta[:]).EncodeToBuffer(buffer[offset:])
	offset += codec.Bool(this.deltaSign).EncodeToBuffer(buffer[offset:])
	return offset
}

func (this *U256) Decode(buffer []byte) interface{} {
	this = &U256{
		finalized: bool(codec.Bool(true).Decode(buffer).(codec.Bool)),
		value:     uint256.NewInt(0),
		min:       uint256.NewInt(0),
		max:       uint256.NewInt(0),
		delta:     uint256.NewInt(0),
		deltaSign: true, // negative
	}

	copy(this.value[:], codec.Uint64s{}.Decode(buffer[1+32*0:1+32*1]).(codec.Uint64s))
	copy(this.min[:], codec.Uint64s{}.Decode(buffer[1+32*1:1+32*2]).(codec.Uint64s))
	copy(this.max[:], codec.Uint64s{}.Decode(buffer[1+32*2:1+32*3]).(codec.Uint64s))
	copy(this.delta[:], codec.Uint64s{}.Decode(buffer[1+32*3:1+32*4]).(codec.Uint64s))
	this.deltaSign = bool(codec.Bool(true).Decode(buffer[1+32*4 : 1+32*4+1]).(codec.Bool))

	return this
}

func (this *U256) EncodeCompact() []byte {
	totalLen := 32
	if this.min != nil {
		totalLen += 32
	}

	if this.max != nil {
		totalLen += 32
	}

	buffer := make([]byte, 2+totalLen) // labels + actual length

	offset := codec.Uint64s(this.value[:]).EncodeToBuffer(buffer[2:])
	if this.min != nil {
		buffer[0] = 1
		offset += codec.Uint64s(this.min[:]).EncodeToBuffer(buffer[offset+2:])
	}

	if this.max != nil {
		buffer[1] = 1
		codec.Uint64s(this.max[:]).EncodeToBuffer(buffer[offset+2:])
	}
	return buffer
}

func (this *U256) DecodeCompact(buffer []byte) interface{} {
	v := uint256.NewInt(0)
	offset := 2
	copy((*v)[:], codec.Uint64s{}.Decode(buffer[offset:offset+32]).(codec.Uint64s))

	var min, max *uint256.Int
	offset += 32
	if buffer[0] == 1 {
		min = uint256.NewInt(0)
		copy(min[:], codec.Uint64s{}.Decode(buffer[offset:offset+32]).(codec.Uint64s))
		offset += 32
	}

	if buffer[1] == 1 {
		max = uint256.NewInt(0)
		copy(max[:], codec.Uint64s{}.Decode(buffer[offset:offset+32]).(codec.Uint64s))
	}

	return &U256{
		value:     v,
		delta:     uint256.NewInt(0),
		min:       min,
		max:       max,
		deltaSign: this.deltaSign,
	}
}
