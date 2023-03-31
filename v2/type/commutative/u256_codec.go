package commutative

import (
	"math/big"

	codec "github.com/arcology-network/common-lib/codec"
	uint256 "github.com/holiman/uint256"
)

func (this *U256) HeaderSize() uint32 {
	return (1 + 5) * codec.UINT32_LEN // Total number of fields + offsets of these fields
}

func (this *U256) Size() uint32 {
	delta := codec.Bigint(*this.delta)
	return codec.Bool(this.finalized).Size() +
		32 + // Values
		32 + // Min
		32 + // Max
		delta.Size()
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

	delta := codec.Bigint(*this.delta)
	return offset + (&delta).EncodeToBuffer(buffer[offset:])
}

func (*U256) Decode(buffer []byte) interface{} {
	balance := &U256{
		finalized: bool(codec.Bool(true).Decode(buffer).(codec.Bool)),
		value:     uint256.NewInt(0),
		min:       uint256.NewInt(0),
		max:       uint256.NewInt(0),
	}

	copy(balance.value[:], codec.Uint64s{}.Decode(buffer[1+32*0:1+32*1]).(codec.Uint64s))
	copy(balance.min[:], codec.Uint64s{}.Decode(buffer[1+32*1:1+32*2]).(codec.Uint64s))
	copy(balance.max[:], codec.Uint64s{}.Decode(buffer[1+32*2:1+32*3]).(codec.Uint64s))
	balance.delta = (*big.Int)((&codec.Bigint{}).Decode(buffer[1+32*3:]).(*codec.Bigint))
	return balance
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
		value: v,
		delta: big.NewInt(0),
		min:   min,
		max:   max,
	}
}
