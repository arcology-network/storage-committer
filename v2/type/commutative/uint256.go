package commutative

import (
	"errors"
	"fmt"

	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	uint256 "github.com/holiman/uint256"
)

const (
	UNKNOWN = iota
	ADDITION
	SUBTRACT
	MULTIPLY
	DIVIDE
)

var (
	U256MIN = uint256.NewInt(0) // default limits
	U256MAX = uint256.NewInt(0).SetAllOne()

	UINT256ZERO = uint256.NewInt(0)
	UINT256ONE  = uint256.NewInt(1)
)

type U256 struct {
	finalized bool // For commutative values only
	value     *uint256.Int
	delta     *uint256.Int
	min       *uint256.Int
	max       *uint256.Int
	operation uint8
}

func (this *U256) IsSelf(key interface{}) bool { return true }
func (this *U256) Composite() bool             { return !this.finalized }
func (this *U256) TypeID() uint8               { return ccurlcommon.CommutativeUint256 }

func NewU256(value, delta, min, max *uint256.Int, operation uint8) interface{} {
	if value.Cmp(min) == -1 || value.Cmp(max) == 1 || max.Cmp(min) == -1 {
		return nil
	}

	return &U256{
		value:     value,
		delta:     delta,
		min:       min,
		max:       max,
		operation: operation,
	}
}

func NewU256Delta(delta *uint256.Int) interface{} {
	return &U256{
		value:     uint256.NewInt(0),
		delta:     delta,
		min:       U256MIN,
		max:       U256MAX,
		operation: UNKNOWN,
	}
}

func (this *U256) FromBytes(value, min, max []byte) {
	this.min.SetBytes(value)
	this.min.SetBytes(min)
	this.max.SetBytes(max)
}

func (this *U256) HasCustomizedLimit() bool {
	return this.min.Cmp(U256MIN) != 0 || this.max.Cmp(U256MAX) != 0
}

func (this *U256) Deepcopy() interface{} {
	return &U256{
		finalized: this.finalized,
		value:     this.value.Clone(),
		delta:     this.delta.Clone(),
		min:       this.min.Clone(),
		max:       this.max.Clone(),
		operation: this.operation,
	}
}

func (this *U256) Value() interface{} {
	return this.value
}

func (this *U256) ToAccess() interface{} {
	return this
}

func (this *U256) checkLimits(value *uint256.Int, delta *uint256.Int) (*uint256.Int, bool) {
	switch this.operation {
	case ADDITION:
		return value.Clone().AddOverflow(value, delta)
	case SUBTRACT:
		return value.Clone().SubOverflow(value, delta)
	case MULTIPLY:
		return value.Clone().MulDivOverflow(value, delta, uint256.NewInt(1))
	case DIVIDE:
		return value.Clone().MulDivOverflow(value, uint256.NewInt(1), delta)
	}
	return nil, false
}

func (this *U256) Get(path string, source interface{}) (interface{}, uint32, uint32) {
	this.finalized = true
	temp := &U256{
		finalized: this.finalized,
		value:     this.value.Clone(),
		min:       this.min,
		max:       this.max,
		operation: this.operation,
	}

	if this.isReadEquivalent(this.delta) {
		return temp, 1, 0
	}

	temp.delta = this.delta.Clone()

	switch this.operation {
	case ADDITION:
		temp.value = this.value.Add(temp.value, temp.delta)
	case SUBTRACT:
		temp.value = this.value.Sub(temp.value, temp.delta)
	case MULTIPLY:
		temp.value = this.value.Mul(temp.value, temp.delta)
	case DIVIDE:
		temp.value = this.value.Div(temp.value, temp.delta)
	}

	return temp, 1, 1
}

func (this *U256) Delta() interface{} {
	return this
}

// Set delta
func (this *U256) Set(path string, newVal interface{}, source interface{}) (uint32, uint32, error) {
	delta := newVal.(*U256).delta
	if this.isReadEquivalent(delta) {
		return 1, 0, nil
	}

	if accumDelta, outOfRange := this.checkLimits(this.delta, delta); !outOfRange {
		if tempV, outOfRange := this.checkLimits(this.value, accumDelta); !outOfRange {
			if this.min.Cmp(tempV) < 1 && tempV.Cmp(this.max) < 1 {
				this.delta = accumDelta
				return 0, 1, nil
			}
			return 0, 1, errors.New("Error: Value out of range")
		}
	}
	return 0, 1, errors.New("Error: Value out of range")
}

func (this *U256) isReadEquivalent(delta *uint256.Int) bool {
	if (this.operation == ADDITION || this.operation == SUBTRACT) && delta.Eq(UINT256ZERO) { // Is equal to a read
		return true
	}

	if (this.operation == MULTIPLY || this.operation == DIVIDE) && delta.Eq(UINT256ONE) {
		return true
	}
	return false
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
