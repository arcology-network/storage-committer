package commutative

import (
	"errors"

	codec "github.com/arcology-network/common-lib/codec"
	uint256 "github.com/holiman/uint256"
)

type U256Delta struct {
	value         *uint256.Int
	deltaPositive bool
}

func (this *U256Delta) isOverflowed(v0 *uint256.Int, signV0 bool, v1 *uint256.Int, signV1 bool) (*uint256.Int, bool) {
	if signV0 == signV1 { // Both positive or negative
		summed, overflowed := v0.AddOverflow(v0, v1)
		if overflowed {
			return nil, true
		}
		return summed, signV0
	}

	if v0.Cmp(v1) < 1 { // v0 <= v1
		return uint256.NewInt(0).Sub(v1, v0), signV1
	}
	return uint256.NewInt(0).Sub(v0, v1), signV0
}

func (this *U256Delta) Add(newDelta interface{}, sign bool) (interface{}, error) {
	accumDelta, deltaSign := this.isOverflowed(this.value.Clone(), this.deltaPositive, (*uint256.Int)(newDelta.(*U256).value), newDelta.(*U256).deltaPositive)
	if accumDelta == nil {
		return this, errors.New("Error: Value out of range")
	}

	tempV, possitive := this.isOverflowed(this.value.Clone(), true, accumDelta.Clone(), deltaSign)
	if tempV == nil || !possitive { // Result must be possitive
		return this, errors.New("Error: Value out of range")
	}
	return this, nil
}

func (this *U256Delta) Size() uint32 { return 32 + 1 }

func (this *U256Delta) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *U256Delta) EncodeToBuffer(buffer []byte) int {
	offset := codec.Bytes(this.value.Bytes()).EncodeToBuffer(buffer)
	offset += codec.Bool(this.deltaPositive).EncodeToBuffer(buffer[offset:])
	return offset
}

func (this *U256Delta) Decode(buffer []byte) interface{} {
	this = &U256Delta{
		new(uint256.Int).SetBytes(buffer),
		bool(codec.Bool(true).Decode(buffer[32:]).(codec.Bool)),
	}
	return this
}
