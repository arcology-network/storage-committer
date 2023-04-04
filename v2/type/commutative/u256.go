package commutative

import (
	"errors"
	"fmt"
	"math/big"

	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	uint256 "github.com/holiman/uint256"
)

var (
	uint256min = uint256.NewInt(0)
	uint256max = uint256.NewInt(0).SetAllOne()
)

type U256 struct {
	finalized bool
	value     *uint256.Int
	min       *uint256.Int
	max       *uint256.Int
	delta     *big.Int
}

func NewU256(value *uint256.Int, delta *big.Int) interface{} {
	return &U256{
		value: value,
		delta: delta,
		min:   uint256min.Clone(),
		max:   uint256max.Clone(),
	}
}

func (this *U256) IsSelf(key interface{}) bool { return true }
func (this *U256) Composite() bool             { return !this.finalized }
func (this *U256) TypeID() uint8               { return ccurlcommon.CommutativeUint256 }

func NewU256WithLimit(min, max *uint256.Int) interface{} {
	return &U256{
		min: min,
		max: max,
	}
}

func (this *U256) HasCustomizedLimit() bool {
	return this.min.Cmp(uint256min) != 0 || this.max.Cmp(uint256max) != 0
}

func (this *U256) Deepcopy() interface{} {
	if this.value == nil {
		this.value = uint256.NewInt(0)
	}
	return &U256{
		this.finalized,
		this.value.Clone(),
		this.min.Clone(),
		this.max.Clone(),
		new(big.Int).Set(this.delta),
	}
}

func (this *U256) Value() interface{} {
	return this.value
}

func (this *U256) ToAccess() interface{} {
	return this
}

func (*U256) checkLimits(value *uint256.Int, deltaBigInt *big.Int, min, max *uint256.Int) (bool, *uint256.Int, error) {
	b := new(big.Int).Set(deltaBigInt)
	delta, failed := uint256.FromBig(b.Abs(b))
	if failed {
		return false, nil, errors.New("Error: Delta Overflow!!!")
	}

	isNegative := deltaBigInt.Sign() == -1
	if isNegative {
		if diff, overflow := new(uint256.Int).SubOverflow(value, delta); overflow {
			return isNegative, nil, errors.New("Error: Underflow!!!")
		} else {
			if min != nil && diff.Cmp(min) == -1 {
				return isNegative, nil, errors.New("Error: Sum overflow!!!")
			}
		}
	} else {
		if sum, overflow := new(uint256.Int).AddOverflow(value, delta); overflow {
			return isNegative, nil, errors.New("Error: Sum overflow!!!")
		} else {
			if max != nil && sum.Cmp(max) == 1 {
				return isNegative, nil, errors.New("Error: Sum overflow!!!")
			}
		}
	}

	return isNegative, delta, nil
}

func (this *U256) Get(path string, source interface{}) (interface{}, uint32, uint32) {
	if this.delta == nil || this.delta.Cmp(big.NewInt(0)) == 0 {
		return this, 1, 0
	}

	this.finalized = true
	temp := &U256{
		finalized: this.finalized,
		value:     this.value.Clone(),
		min:       this.min,
		max:       this.max,
		delta:     big.NewInt(0),
	}

	delta := uint256.NewInt(0)
	delta.SetFromBig(this.delta)
	if this.delta.Cmp(big.NewInt(0)) < 0 {
		temp.value.Sub(temp.value, delta)
	} else {
		temp.value.Add(temp.value, delta)
	}
	return temp, 1, 1
}

func (this *U256) Delta() interface{} {
	return this
}

// Set delta
func (this *U256) Set(path string, v interface{}, source interface{}) (uint32, uint32, error) {
	b := v.(*U256)
	if _, _, err := this.checkLimits(this.value, new(big.Int).Add(this.delta, b.delta), this.min, this.max); err != nil {
		return 0, 1, err
	}

	this.delta.Add(this.delta, b.delta)
	return 0, 1, nil
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
		this.delta = big.NewInt(0)
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
			if v.(*U256).Composite() {
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
	this.delta = big.NewInt(0)
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
