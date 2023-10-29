package commutative

import (
	"math/big"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/evm/rlp"
)

func (this *U256) HeaderSize() uint32 {
	return 5 // Total number of fields + offsets of these fields
}

func (this *U256) Size() uint32 {
	return this.HeaderSize() +
		common.IfThen(this.value.Eq(&U256_ZERO), 0, uint32(32)) + // Values
		common.IfThen(this.delta.Eq(&U256_ZERO), 0, uint32(32)) + // delta
		common.IfThen(this.deltaPositive, 0, uint32(1)) + // delta sign
		common.IfThen(this.min.Eq(&U256_ZERO), 0, uint32(32)) + // Min
		common.IfThen(this.max.Eq(&U256_MAX), 0, uint32(32)) // Max
}

func (this *U256) Encode() []byte {
	buffer := make([]byte, this.Size())
	buffer[0] = common.IfThen(this.value.Eq(&U256_ZERO), 0, uint8(32))
	buffer[1] = common.IfThen(this.delta.Eq(&U256_ZERO), 0, uint8(32))
	buffer[2] = common.IfThen(this.deltaPositive, 0, uint8(1)) //only is the delta != 0
	buffer[3] = common.IfThen(this.min.Eq(&U256_ZERO), 0, uint8(32))
	buffer[4] = common.IfThen(this.max.Eq(&U256_MAX), 0, uint8(32))

	this.EncodeToBuffer(buffer[5:])
	return buffer
}

func (this *U256) EncodeToBuffer(buffer []byte) int {
	offset := common.IfThenDo1st(!this.value.Eq(&U256_ZERO), func() int { return codec.Uint64s(this.value[:]).EncodeToBuffer(buffer) }, 0)
	offset += common.IfThenDo1st(!this.delta.Eq(&U256_ZERO), func() int { return codec.Uint64s(this.delta[:]).EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(!this.deltaPositive, func() int { return codec.Bool(this.deltaPositive).EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(!this.min.Eq(&U256_ZERO), func() int { return codec.Uint64s(this.min[:]).EncodeToBuffer(buffer[offset:]) }, 0)
	offset += common.IfThenDo1st(!this.max.Eq(&U256_MAX), func() int { return codec.Uint64s(this.max[:]).EncodeToBuffer(buffer[offset:]) }, 0)
	return offset
}

func (this *U256) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return this
	}
	this = NewUnboundedU256().(*U256)

	offset := 5
	if buffer[0] > 0 {
		copy(this.value[:], codec.Uint64s{}.Decode(buffer[offset:]).(codec.Uint64s))
		offset += int(buffer[0])
	}

	if buffer[1] > 0 {
		copy(this.delta[:], codec.Uint64s{}.Decode(buffer[offset:]).(codec.Uint64s))
		offset += int(buffer[1])
	}

	if buffer[2] > 0 {
		this.deltaPositive = bool(codec.Bool(true).Decode(buffer[offset:]).(codec.Bool))
		offset += int(buffer[2])
	}

	if buffer[3] > 0 {
		copy(this.min[:], codec.Uint64s{}.Decode(buffer[offset:]).(codec.Uint64s))
		offset += int(buffer[3])
	}

	if buffer[4] > 0 {
		copy(this.max[:], codec.Uint64s{}.Decode(buffer[offset:]).(codec.Uint64s))
	}

	return this
}

func (this *U256) StorageEncode() []byte {
	var buffer []byte
	if this.IsBounded() {
		buffer, _ = rlp.EncodeToBytes([]interface{}{this.value, this.min, this.max})
	} else {
		buffer, _ = rlp.EncodeToBytes(this.value.ToBig())
	}
	return buffer
}

func (*U256) StorageDecode(buffer []byte) interface{} {
	this := NewUnboundedU256().(*U256)

	var arr []interface{}
	err := rlp.DecodeBytes(buffer, &arr)
	if err != nil {
		var v2 big.Int
		if err = rlp.DecodeBytes(buffer, &v2); err == nil {
			this.value.SetFromBig(&v2)

		}
	} else {
		this.value.SetBytes(arr[0].([]byte))
		this.min.SetBytes(arr[1].([]byte))
		this.max.SetBytes(arr[2].([]byte))
	}
	return this
}
