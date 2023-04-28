package commutative

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	uint256 "github.com/holiman/uint256"
)

var (
	U256MIN = uint256.NewInt(0) // default limits
	U256MAX = uint256.NewInt(0).SetAllOne()

	UINT256ZERO = uint256.NewInt(0)
	UINT256ONE  = uint256.NewInt(1)
)

type U256 struct {
	value         *uint256.Int
	delta         *uint256.Int
	min           *uint256.Int
	max           *uint256.Int
	deltaPositive bool
}

func NewU256(value, min, max *uint256.Int) interface{} {
	if value.Cmp(min) == -1 || value.Cmp(max) == 1 || max.Cmp(min) == -1 {
		return nil
	}

	return &U256{
		value:         value,
		delta:         uint256.NewInt(0),
		min:           min,
		max:           max,
		deltaPositive: true, // positive delta by default
	}
}

func (this *U256) IsSelf(key interface{}) bool { return true }
func (this *U256) TypeID() uint8               { return ccurlcommon.Commutative{}.Uint256() }
func (this *U256) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

func NewU256FromBytes(value []byte, min, max []byte) interface{} {
	this := &U256{} // positive delta by default
	this.FromBytes(value, min, max)
	return this
}

func NewU256Delta(delta *uint256.Int, deltaPositive bool) interface{} {
	return &U256{
		value:         nil,
		min:           nil,
		max:           nil,
		delta:         delta,
		deltaPositive: deltaPositive,
	}
}

func NewU256DeltaFromBigInt(delta *big.Int) (interface{}, bool) {
	sign := delta.Sign()
	deltaV, overflowed := uint256.FromBig(delta.Abs(delta))
	if overflowed {
		return nil, false
	}

	return &U256{
		delta:         deltaV,
		deltaPositive: sign != -1, // >= 0
	}, true
}

func (this *U256) FromBytes(value []byte, min, max []byte) {
	this.value.SetBytes(value)
	this.min.SetBytes(min)
	this.max.SetBytes(max)
	this.deltaPositive = true
}

func (this *U256) HasCustomizedLimit() bool {
	return this.min.Cmp(U256MIN) != 0 || this.max.Cmp(U256MAX) != 0
}

func (this *U256) Deepcopy() interface{} {
	return &U256{
		value:         this.value.Clone(),
		delta:         this.delta.Clone(),
		min:           this.min.Clone(),
		max:           this.max.Clone(),
		deltaPositive: this.deltaPositive,
	}
}

func (this *U256) Equal(other interface{}) bool {
	return bytes.Equal(this.value.Bytes(), other.(*U256).value.Bytes()) &&
		bytes.Equal(this.delta.Bytes(), other.(*U256).delta.Bytes()) &&
		bytes.Equal(this.min.Bytes(), other.(*U256).min.Bytes()) &&
		bytes.Equal(this.max.Bytes(), other.(*U256).max.Bytes()) &&
		this.deltaPositive == other.(*U256).deltaPositive
}

func (this *U256) Get() (interface{}, uint32, uint32) {
	if this.delta.Eq(UINT256ZERO) {
		return this.value, 1, 0
	}
	return new(uint256.Int).Add(this.value.Clone(), this.delta), 1, 1
}

func (this *U256) Value() interface{} {
	v := &U256{
		delta:         this.delta.Clone(),
		deltaPositive: this.deltaPositive,
	}

	if this.deltaPositive {
		v.value = this.value.Clone().Abs(this.delta)
	} else {
		v.value = this.value.Clone().Sub(this.value, this.delta)
	}
	return this.value
}

func (this *U256) Delta() interface{} { return this }

func (this *U256) isOverflowed(v0 *uint256.Int, signV0 bool, v1 *uint256.Int, signV1 bool) (*uint256.Int, bool) {
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

// Set delta
func (this *U256) Set(newDelta interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if newDelta.(*U256).delta.Eq(UINT256ZERO) {
		return this, 1, 0, 0, nil
	}

	accumDelta, deltaSign := this.isOverflowed(this.delta.Clone(), this.deltaPositive, newDelta.(*U256).delta, newDelta.(*U256).deltaPositive)
	if accumDelta == nil {
		return this, 0, 0, 1, errors.New("Error: Value out of range")
	}

	tempV, possitive := this.isOverflowed(this.value.Clone(), true, accumDelta.Clone(), deltaSign)
	if tempV == nil || !possitive { // Result must be possitive
		return this, 0, 0, 1, errors.New("Error: Value out of range")
	}

	if this.min.Cmp(tempV) < 1 && tempV.Cmp(this.max) < 1 {
		this.delta = accumDelta
		this.deltaPositive = deltaSign
		return this, 0, 0, 1, nil
	}
	return this, 0, 0, 1, errors.New("Error: Value out of range")
}

func (this *U256) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	vec := v.([]ccurlcommon.UnivalueInterface)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*U256)
		}

		if this == nil && v == nil { // Delete a non-existent
			this = nil
		}

		if this != nil && v != nil { // Update an existent
			if _, _, _, _, err := this.Set(v.(*U256), nil); err != nil {
				panic(err)
			}
		}

		if this != nil && v == nil { // Delete an existent
			this = nil
		}
	}

	newValue, _, _ := this.Get()
	this.value = newValue.(*uint256.Int)
	this.delta.Clear()
	return this
}

func (this *U256) Purge() {
	this.delta = uint256.NewInt(0)
}

func (this *U256) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

func (this *U256) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}

func (this *U256) GetDelta() interface{} {
	return this.delta
}
