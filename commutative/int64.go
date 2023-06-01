package commutative

import (
	"errors"
	"math"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type Int64 struct {
	value *codec.Int64
	delta *codec.Int64
	min   *codec.Int64
	max   *codec.Int64
}

func NewInt64(min, max int64) interface{} {
	if min > max {
		return nil
	}

	return &Int64{
		common.New(codec.Int64(0)),
		common.New(codec.Int64(0)),
		common.New(codec.Int64(min)),
		common.New(codec.Int64(max)),
	}
}

func NewInt64Delta(delta int64) interface{} {
	deltaV := codec.Int64(delta)
	return &Int64{delta: &deltaV}
}

func (this *Int64) New(value, delta, sign, min, max interface{}) interface{} {
	return &Int64{
		common.IfThenDo1st(value != nil, func() *codec.Int64 { return value.(*codec.Int64) }, common.New(codec.Int64(0))),
		common.IfThenDo1st(delta != nil, func() *codec.Int64 { return delta.(*codec.Int64) }, common.New(codec.Int64(0))),
		common.IfThenDo1st(min != nil, func() *codec.Int64 { return min.(*codec.Int64) }, common.New(codec.Int64(math.MinInt64))),
		common.IfThenDo1st(max != nil, func() *codec.Int64 { return max.(*codec.Int64) }, common.New(codec.Int64(math.MaxInt64))),
	}
}

func (this *Int64) Clone() interface{} {
	return &Int64{
		value: common.New(*this.value),
		delta: common.New(*this.delta),
		min:   common.New(*this.min),
		max:   common.New(*this.max),
	}
}

func (this *Int64) Equal(other interface{}) bool {
	return common.Equal(this.value, other.(*Int64).value, func(v *codec.Int64) bool { return *v == 0 }) &&
		common.Equal(this.delta, other.(*Int64).delta, func(v *codec.Int64) bool { return *v == 0 }) &&
		common.Equal(this.min, other.(*Int64).min, func(v *codec.Int64) bool { return *v == math.MinInt64 }) &&
		common.Equal(this.max, other.(*Int64).max, func(v *codec.Int64) bool { return *v == math.MaxInt64 })
}

func (this *Int64) IsNumeric() bool     { return true }
func (this *Int64) IsCommutative() bool { return true }

func (this *Int64) Value() interface{} { return (this.value) }
func (this *Int64) Delta() interface{} { return (this.delta) }
func (this *Int64) DeltaSign() bool    { return *this.delta >= 0 }
func (this *Int64) Min() interface{}   { return (this.min) }
func (this *Int64) Max() interface{}   { return (this.max) }

func (this *Int64) SetValue(v interface{})     { *this.value = *(v.(*codec.Int64)) }
func (this *Int64) SetDelta(v interface{})     { *this.delta = *(v.(*codec.Int64)) }
func (this *Int64) SetDeltaSign(v interface{}) {}
func (this *Int64) SetMin(v interface{})       { *this.min = *(v.(*codec.Int64)) }
func (this *Int64) SetMax(v interface{})       { *this.max = *(v.(*codec.Int64)) }

func (this *Int64) MemSize() uint32                                            { return 5 * 8 }
func (this *Int64) TypeID() uint8                                              { return INT64 }
func (this *Int64) IsSelf(key interface{}) bool                                { return true }
func (this *Int64) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *Int64) Get() (interface{}, uint32, uint32) {
	return int64(*this.value + *this.delta), 1, common.IfThen(*this.delta == 0, uint32(0), uint32(1))
}

func (this *Int64) Set(v interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if this.isUnderflow(int64(*v.(*Int64).delta)) || this.isOverflow(int64(*v.(*Int64).delta)) {
		return this, 0, 1, 0, errors.New("Error: Value out of range!!")
	}

	if (*this.min > *this.value+*this.delta+*v.(*Int64).delta) ||
		(*this.max < *this.value+*this.delta+*v.(*Int64).delta) {
		return this, 0, 1, 0, errors.New("Error: Value out of range!!")
	}

	*this.delta += *v.(*Int64).delta
	return this, 0, 1, 0, nil
}

func (this *Int64) isOverflow(delta int64) bool {
	flag := *this.max-codec.Int64(delta) < *this.value+*this.delta
	return (delta >= 0 && *this.delta >= 0) &&
		(*this.max < codec.Int64(delta) || flag)
}

func (this *Int64) isUnderflow(delta int64) bool {
	flag := *this.min-codec.Int64(delta) > *this.value+*this.delta
	return (delta < 0 && *this.delta < 0) &&
		(*this.min > codec.Int64(delta) || flag)
}

func (this *Int64) ApplyDelta(v interface{}) (interfaces.Type, int, error) {
	vec := v.([]interfaces.Univalue)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Int64)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			if _, _, _, _, err := this.Set(v.(*Int64), nil); err != nil {
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

func (this *Int64) Reset() {
	*this.delta = 0
}

func (this *Int64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}