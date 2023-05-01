package commutative

import (
	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	uint256 "github.com/holiman/uint256"
)

func (this *U256) HeaderSize() uint32 {
	return (1 + 5) * codec.UINT32_LEN // Total number of fields + offsets of these fields
}

func (this *U256) Size(selectors ...interface{}) uint32 {
	return this.HeaderSize() +
		common.IfThen(len(selectors) == 0 || selectors[0].(bool), uint32(32), 0) + // Values
		common.IfThen(len(selectors) == 0 || selectors[1].(bool), uint32(32), 0) + // delta
		common.IfThen(len(selectors) == 0 || selectors[1].(bool), uint32(1), 0) + // delta sign
		common.IfThen(len(selectors) == 0 || selectors[2].(bool), uint32(32), 0) + // Min
		common.IfThen(len(selectors) == 0 || selectors[3].(bool), uint32(32), 0) // Max
}

func (this *U256) Encode(selectors ...interface{}) []byte {
	buffer := make([]byte, this.Size(selectors...))
	offset := codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			common.IfThen(len(selectors) == 0 || selectors[0].(bool), uint32(32), 0),
			common.IfThen(len(selectors) == 0 || selectors[1].(bool), uint32(32), 0),
			common.IfThen(len(selectors) == 0 || selectors[1].(bool), uint32(1), 0),
			common.IfThen(len(selectors) == 0 || selectors[2].(bool), uint32(32), 0),
			common.IfThen(len(selectors) == 0 || selectors[3].(bool), uint32(32), 0),
		},
	)
	this.EncodeToBuffer(buffer[offset:], selectors...)
	return buffer
}

func (this *U256) EncodeToBuffer(buffer []byte, selectors ...interface{}) int {
	offset := common.IfThenDo1st(len(selectors) == 0 || selectors[0].(bool), func() int { return codec.Uint64s(this.value[:]).EncodeToBuffer(buffer) }, 0)
	offset += common.IfThenDo1st(len(selectors) == 0 || selectors[1].(bool), func() int { return codec.Uint64s(this.delta[:]).EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(len(selectors) == 0 || selectors[1].(bool), func() int { return codec.Bool(this.deltaPositive).EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(len(selectors) == 0 || selectors[2].(bool), func() int { return codec.Uint64s(this.min[:]).EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(len(selectors) == 0 || selectors[3].(bool), func() int { return codec.Uint64s(this.max[:]).EncodeToBuffer(buffer[offset:]) }, 0)
	return offset
}

func (this *U256) Decode(buffer []byte) interface{} {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)

	this = &U256{
		value:         uint256.NewInt(0),
		delta:         uint256.NewInt(0),
		deltaPositive: true,
		min:           uint256.NewInt(0),
		max:           new(uint256.Int).SetBytes(common.Fill(make([]byte, 32), 255)),
	}

	copy(this.value[:], codec.Uint64s{}.Decode(fields[0]).(codec.Uint64s))
	copy(this.delta[:], codec.Uint64s{}.Decode(fields[1]).(codec.Uint64s))
	this.deltaPositive = bool(codec.Bool(true).Decode(fields[2]).(codec.Bool))
	copy(this.min[:], codec.Uint64s{}.Decode(fields[3]).(codec.Uint64s))
	copy(this.max[:], codec.Uint64s{}.Decode(fields[4]).(codec.Uint64s))

	return this
}

// func (this *U256) Encode() []byte {
// 	totalLen := 32
// 	if this.min != nil {
// 		totalLen += 32
// 	}

// 	if this.max != nil {
// 		totalLen += 32
// 	}

// 	buffer := make([]byte, 2+totalLen) // labels + actual length

// 	offset := codec.Uint64s(this.value[:]).EncodeToBuffer(buffer[2:])
// 	if this.min != nil {
// 		buffer[0] = 1
// 		offset += codec.Uint64s(this.min[:]).EncodeToBuffer(buffer[offset+2:])
// 	}

// 	if this.max != nil {
// 		buffer[1] = 1
// 		codec.Uint64s(this.max[:]).EncodeToBuffer(buffer[offset+2:])
// 	}
// 	return buffer
// }

// func (this *U256) DecodeCompact(buffer []byte) interface{} {
// 	v := uint256.NewInt(0)
// 	offset := 2
// 	copy((*v)[:], codec.Uint64s{}.Decode(buffer[offset:offset+32]).(codec.Uint64s))

// 	var min, max *uint256.Int
// 	offset += 32
// 	if buffer[0] == 1 {
// 		min = uint256.NewInt(0)
// 		copy(min[:], codec.Uint64s{}.Decode(buffer[offset:offset+32]).(codec.Uint64s))
// 		offset += 32
// 	}

// 	if buffer[1] == 1 {
// 		max = uint256.NewInt(0)
// 		copy(max[:], codec.Uint64s{}.Decode(buffer[offset:offset+32]).(codec.Uint64s))
// 	}

// 	return &U256{
// 		value:         v,
// 		delta:         uint256.NewInt(0),
// 		min:           min,
// 		max:           max,
// 		deltaPositive: this.deltaPositive,
// 	}
// }
