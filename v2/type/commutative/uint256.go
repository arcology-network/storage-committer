package commutative

import (
	"errors"
	"fmt"
	"math/big"

	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	uint256 "github.com/holiman/uint256"
)

const (
	UNKNOWN = iota
	ADDITION
	SUBTRACT
	MULTIPLY
	DIVIDE

	ADD_SUB
	MUL_DIV
)

var (
	U256MIN = uint256.NewInt(0) // default limits
	U256MAX = uint256.NewInt(0).SetAllOne()

	UINT256ZERO = uint256.NewInt(0)
	UINT256ONE  = uint256.NewInt(1)
)

type U256 struct {
	finalized      bool // For commutative values only
	value          *uint256.Int
	delta          *uint256.Int
	min            *uint256.Int
	max            *uint256.Int
	deltaPossitive bool
}

func (this *U256) IsSelf(key interface{}) bool { return true }
func (this *U256) DeltaWritable() bool         { return !this.finalized }
func (this *U256) TypeID() uint8               { return ccurlcommon.CommutativeUint256 }

func NewU256(value, min, max *uint256.Int) interface{} {
	if value.Cmp(min) == -1 || value.Cmp(max) == 1 || max.Cmp(min) == -1 {
		return nil
	}

	return &U256{
		value:          value,
		delta:          uint256.NewInt(0),
		min:            min,
		max:            max,
		deltaPossitive: true, // positive delta by default
	}
}

func NewU256FromBytes(value []byte, min, max []byte) interface{} {
	this := &U256{} // positive delta by default
	this.FromBytes(value, min, max)
	return this
}

func NewU256Delta(delta *uint256.Int, deltaPossitive bool) interface{} {
	return &U256{
		value:          nil,
		min:            nil,
		max:            nil,
		delta:          delta,
		deltaPossitive: deltaPossitive,
	}
}

func NewU256DeltaFromBigInt(delta *big.Int) (interface{}, bool) {
	sign := delta.Sign()
	deltaV, overflowed := uint256.FromBig(delta.Abs(delta))
	if overflowed {
		return nil, false
	}

	return &U256{
		delta:          deltaV,
		deltaPossitive: sign != -1, // >= 0
	}, true
}

func (this *U256) FromBytes(value []byte, min, max []byte) {
	this.value.SetBytes(value)
	this.min.SetBytes(min)
	this.max.SetBytes(max)
	this.deltaPossitive = true
}

func (this *U256) HasCustomizedLimit() bool {
	return this.min.Cmp(U256MIN) != 0 || this.max.Cmp(U256MAX) != 0
}

func (this *U256) Deepcopy() interface{} {
	return &U256{
		finalized:      this.finalized,
		value:          this.value.Clone(),
		delta:          this.delta.Clone(),
		min:            this.min.Clone(),
		max:            this.max.Clone(),
		deltaPossitive: this.deltaPossitive,
	}
}

func (this *U256) Value() interface{} {
	return this.value
}

func (this *U256) ToAccess() interface{} {
	return this
}

func (this *U256) isOverflowed(v0 *uint256.Int, signV0 bool, v1 *uint256.Int, signV1 bool) (*uint256.Int, bool) {
	if signV0 == signV1 { // Possitive
		v, overflowed := v0.Clone().AddOverflow(v0, v1)
		if overflowed {
			return nil, signV0
		}
		return v, signV0
	}

	if v0.Cmp(v1) < 1 { // v0 <= v1
		return v1.Sub(v1, v0), signV1
	}
	return v1.Sub(v0, v1), signV0
}

func (this *U256) Get(path string, source interface{}) (interface{}, uint32, uint32) {
	this.finalized = true
	temp := &U256{
		finalized:      this.finalized,
		value:          this.value.Clone(),
		delta:          this.delta.Clone(),
		min:            this.min,
		max:            this.max,
		deltaPossitive: this.deltaPossitive,
	}

	if this.delta.Eq(UINT256ZERO) {
		return temp, 1, 0
	}

	temp.value.Add(temp.value, temp.delta)
	temp.deltaPossitive = false
	temp.delta.Clear()

	return temp, 1, 1
}

func (this *U256) Delta() interface{} {
	return this
}

// Set delta
func (this *U256) Set(path string, newDelta interface{}, source interface{}) (uint32, uint32, error) {
	if newDelta.(*U256).delta.Eq(UINT256ZERO) {
		return 1, 0, nil
	}

	accumDelta, accumSign := this.isOverflowed(this.delta.Clone(), this.deltaPossitive, newDelta.(*U256).delta, newDelta.(*U256).deltaPossitive)
	if accumDelta == nil {
		return 0, 1, errors.New("Error: Value out of range")
	}

	tempV, deltaPossitive := this.isOverflowed(this.value.Clone(), true, accumDelta.Clone(), accumSign)
	if tempV == nil || !deltaPossitive {
		return 0, 1, errors.New("Error: Value out of range")
	}

	if this.min.Cmp(tempV) < 1 && tempV.Cmp(this.max) < 1 {
		this.delta = accumDelta
		this.deltaPossitive = deltaPossitive
		return 0, 1, nil
	}
	return 0, 1, errors.New("Error: Value out of range")
}

func (this *U256) Reset(path string, v interface{}, source interface{}) (uint32, uint32, error) {
	this.finalized = true
	b := v.(*U256)
	if b.value != nil {
		this.value = b.value
	}

	if this.value == nil {
		this.value = uint256.NewInt(0)
	}

	if b.delta != nil {
		this.delta = b.delta
	} else {
		this.delta = uint256.NewInt(0)
	}

	if b.min != nil {
		this.min = b.min
	}
	if b.max != nil {
		this.max = b.max
	}

	return 0, 1, nil
}

func (this *U256) This(source interface{}) interface{} {
	v, _, _ := this.Deepcopy().(*U256).Get("", source)
	return v
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
			if v.(*U256).DeltaWritable() {
				if _, _, err := this.Set("", v.(*U256), nil); err != nil {
					panic(err)
				}
			} else {
				if _, _, err := this.Reset("", v.(*U256), nil); err != nil {
					panic(err)
				}
			}
		}

		if this != nil && v == nil { // Delete an existent
			this = nil
		}
	}

	newValue, _, _ := this.Get("", nil)
	*this = (*newValue.(*U256))
	return this
}

func (this *U256) Purge() {
	this.finalized = false
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
