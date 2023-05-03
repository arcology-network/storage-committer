package commutative

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	uint256 "github.com/holiman/uint256"
)

var (
	U256_MIN = uint256.NewInt(0) // default limits
	U256_MAX = uint256.NewInt(0).SetAllOne()

	U256_ZERO = uint256.NewInt(0)
	U256_ONE  = uint256.NewInt(1)
)

type U256 struct {
	value         *uint256.Int
	delta         *uint256.Int
	min           *uint256.Int
	max           *uint256.Int
	deltaPositive bool
}

func NewU256(min, max *uint256.Int) interface{} {
	return &U256{
		value:         uint256.NewInt(0),
		delta:         uint256.NewInt(0),
		min:           common.IfThen(min != nil, min.Clone(), U256_MIN.Clone()),
		max:           common.IfThen(max != nil, max.Clone(), U256_MAX.Clone()),
		deltaPositive: true, // positive delta by default
	}
}

func NewU256FromBytes(value []byte, min, max []byte) interface{} {
	this := &U256{
		value:         uint256.NewInt(0),
		delta:         uint256.NewInt(0),
		deltaPositive: true,
	}
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
		value:         nil,
		min:           nil,
		max:           nil,
		delta:         deltaV,
		deltaPositive: sign != -1, // >= 0
	}, true
}

// For the codec only, don't use it for other purposes
func (this *U256) New(value, delta, sign, min, max interface{}) interface{} {
	return &U256{
		value:         common.IfThenDo1st(value != nil && value.(*uint256.Int) != nil && !value.(*uint256.Int).Eq(U256_ZERO.Clone()), func() *uint256.Int { return value.(*uint256.Int) }, nil),
		delta:         common.IfThenDo1st(delta != nil && delta.(*uint256.Int) != nil && !delta.(*uint256.Int).Eq(U256_ZERO.Clone()), func() *uint256.Int { return delta.(*uint256.Int) }, nil),
		deltaPositive: common.IfThenDo1st(sign != nil, func() bool { return sign.(bool) }, true),
		min:           common.IfThenDo1st(min != nil && min.(*uint256.Int) != nil && !min.(*uint256.Int).Eq(U256_MIN.Clone()), func() *uint256.Int { return min.(*uint256.Int) }, nil),
		max:           common.IfThenDo1st(max != nil && max.(*uint256.Int) != nil && !max.(*uint256.Int).Eq(U256_MAX.Clone()), func() *uint256.Int { return max.(*uint256.Int) }, nil),
	}
}

// ReInit the fields with default values that were removed before export.
func (this *U256) ReInit() {
	this.value = common.IfThen(this.value == nil, U256_ZERO.Clone(), this.value)
	this.delta = common.IfThen(this.delta == nil, U256_ZERO.Clone(), this.delta)
	this.deltaPositive = common.IfThen(this.delta == nil, true, this.deltaPositive)
	this.min = common.IfThen(this.min == nil, U256_ZERO.Clone(), this.min)
	this.max = common.IfThen(this.max == nil, U256_MAX.Clone(), this.max)
}

func (this *U256) Value() interface{} {

	// (this.value)
	return this.value
}
func (this *U256) Delta() interface{} { return this.delta }
func (this *U256) Sign() bool         { return this.delta.Cmp(U256_ZERO) >= 0 }
func (this *U256) Min() interface{}   { return this.min }
func (this *U256) Max() interface{}   { return this.max }

func (this *U256) MemSize() uint32                                            { return 32*4 + 1 }
func (this *U256) IsSelf(key interface{}) bool                                { return true }
func (this *U256) TypeID() uint8                                              { return UINT256 }
func (this *U256) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *U256) FromBytes(value []byte, min, max []byte) {
	this.value.SetBytes(value)
	this.min.SetBytes(min)
	this.max.SetBytes(max)
	this.deltaPositive = true
}

func (this *U256) Clone() interface{} {
	return &U256{
		value:         common.IfThenDo1st(this.value != nil, func() *uint256.Int { return this.value.Clone() }, nil),
		delta:         common.IfThenDo1st(this.delta != nil, func() *uint256.Int { return this.delta.Clone() }, nil),
		min:           common.IfThenDo1st(this.min != nil, func() *uint256.Int { return this.min.Clone() }, nil),
		max:           common.IfThenDo1st(this.max != nil, func() *uint256.Int { return this.max.Clone() }, nil),
		deltaPositive: this.deltaPositive,
	}
}

func (this *U256) Equal(other interface{}) bool {
	return common.Equal(this.value, other.(*U256).value, func(v *uint256.Int) bool { return v.Eq(U256_ZERO) }) &&
		common.Equal(this.delta, other.(*U256).delta, func(v *uint256.Int) bool { return v.Eq(U256_ZERO) }) &&
		common.Equal(this.min, other.(*U256).min, func(v *uint256.Int) bool { return v.Eq(U256_ZERO) }) &&
		common.Equal(this.max, other.(*U256).max, func(v *uint256.Int) bool { return v.Eq(U256_MAX) })
}

func (this *U256) Get() (interface{}, uint32, uint32) {
	if this.delta.Eq(U256_ZERO) {
		return this.value, 1, 0
	}
	return new(uint256.Int).Add(this.value.Clone(), this.delta), 1, 1
}

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
	if newDelta.(*U256).delta.Eq(U256_ZERO) {
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
	this.ReInit()
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
	return hasher(this.Encode())
}

func (this *U256) Print() {
	fmt.Println("Value: ", this.value)
	fmt.Println("Delta: ", this.delta)
	fmt.Println()
}
