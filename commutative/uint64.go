package commutative

import (
	"errors"
	"math"

	codec "github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type Selector []bool

type Uint64 struct {
	value *codec.Uint64
	delta *codec.Uint64
	min   *codec.Uint64
	max   *codec.Uint64
}

func NewUint64(limits ...uint64) interface{} {
	limits = common.IfThen(len(limits) == 0, []uint64{0, math.MaxUint64}, limits)
	if limits[0] > limits[1] {
		return nil
	}

	return &Uint64{
		common.New(codec.Uint64(0)),
		common.New(codec.Uint64(0)),
		common.New(codec.Uint64(limits[0])), // min
		common.New(codec.Uint64(limits[1])), // max
	}
}

func NewUint64Delta(delta uint64) interface{} {
	deltaV := codec.Uint64(delta)
	return &Uint64{delta: &deltaV}
}

// For the codec only, don't use it for other purposes
func (this *Uint64) New(value, delta, sign, min, max interface{}) interface{} {
	return &Uint64{
		common.IfThenDo1st(value != nil, func() *codec.Uint64 { return value.(*codec.Uint64) }, common.New(codec.Uint64(0))),
		common.IfThenDo1st(delta != nil, func() *codec.Uint64 { return delta.(*codec.Uint64) }, common.New(codec.Uint64(0))),
		common.IfThenDo1st(min != nil, func() *codec.Uint64 { return min.(*codec.Uint64) }, common.New(codec.Uint64(0))),
		common.IfThenDo1st(max != nil, func() *codec.Uint64 { return max.(*codec.Uint64) }, common.New(codec.Uint64(math.MaxUint64))),
	}
}

func (this *Uint64) Clone() interface{} {
	return &Uint64{
		value: common.New(*this.value),
		delta: common.New(*this.delta),
		min:   common.New(*this.min),
		max:   common.New(*this.max),
	}
}

func (this *Uint64) Equal(other interface{}) bool {
	return common.Equal(this.value, other.(*Uint64).value, func(v *codec.Uint64) bool { return *v == 0 }) &&
		common.Equal(this.delta, other.(*Uint64).delta, func(v *codec.Uint64) bool { return *v == 0 }) &&
		common.Equal(this.min, other.(*Uint64).min, func(v *codec.Uint64) bool { return *v == 0 }) &&
		common.Equal(this.max, other.(*Uint64).max, func(v *codec.Uint64) bool { return *v == math.MaxUint64 })
}

func (this *Uint64) MemSize() uint32 { return 5 * 8 }

func (this *Uint64) IsNumeric() bool     { return true }
func (this *Uint64) IsCommutative() bool { return true }

func (this *Uint64) Value() interface{} { return (this.value) }
func (this *Uint64) Delta() interface{} { return (this.delta) }
func (this *Uint64) DeltaSign() bool    { return true }
func (this *Uint64) Min() interface{}   { return (this.min) }
func (this *Uint64) Max() interface{}   { return (this.max) }

func (this *Uint64) SetValue(v interface{}) { *this.value = *v.(*codec.Uint64) }

func (this *Uint64) ResetDelta()                { this.SetDelta(common.New[codec.Uint64](0)) }
func (this *Uint64) SetDelta(v interface{})     { *this.delta = *v.(*codec.Uint64) }
func (this *Uint64) SetDeltaSign(v interface{}) {}
func (this *Uint64) SetMin(v interface{})       { *this.min = *v.(*codec.Uint64) }
func (this *Uint64) SetMax(v interface{})       { *this.max = *v.(*codec.Uint64) }

func (this *Uint64) TypeID() uint8                                              { return UINT64 }
func (this *Uint64) IsSelf(key interface{}) bool                                { return true }
func (this *Uint64) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *Uint64) Get() (interface{}, uint32, uint32) {
	return uint64(*this.value + *this.delta), 1, common.IfThen(*this.delta == 0, uint32(0), uint32(1))
}

func (this *Uint64) Set(v interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if (*this.max < *v.(*Uint64).delta) || (*this.max-*v.(*Uint64).delta < *this.value+*this.delta) {
		return this, 0, 1, 0, errors.New("Error: Value out of range!!")
	}

	*this.delta += *v.(*Uint64).delta
	return this, 0, 0, 1, nil
}

func (this *Uint64) ApplyDelta(v interface{}) (interfaces.Type, int, error) {
	vec := v.([]interfaces.Univalue)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Uint64)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			if _, _, _, _, err := this.Set(v.(*Uint64), nil); err != nil {
				return nil, i, err
			}
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	if this == nil {
		return nil, 0, errors.New("Error: Nil value")
	}

	*this.value += *this.delta
	*this.delta = 0
	return this, len(vec), nil
}

func (this *Uint64) Reset() {
	*this.delta = 0
}

func (this *Uint64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}
