package commutative

import (
	"errors"
	"math"

	"github.com/arcology-network/common-lib/common"
	intf "github.com/arcology-network/concurrenturl/interfaces"
)

type Int64 struct {
	value int64
	delta int64
	min   int64
	max   int64
}

func NewInt64(min, max int64) interface{} {
	if min > max {
		return nil
	}

	return &Int64{
		0,
		0,
		min,
		max,
	}
}

func NewInt64Delta(delta int64) interface{} {
	return &Int64{delta: delta}
}

func (this *Int64) New(value, delta, sign, min, max interface{}) interface{} {
	return &Int64{
		common.IfThenDo1st(value != nil, func() int64 { return value.(int64) }, 0),
		common.IfThenDo1st(delta != nil, func() int64 { return delta.(int64) }, 0),
		common.IfThenDo1st(min != nil, func() int64 { return min.(int64) }, math.MinInt64),
		common.IfThenDo1st(max != nil, func() int64 { return max.(int64) }, math.MaxInt64),
	}
}

func (this *Int64) Clone() interface{} {
	return &Int64{
		value: this.value,
		delta: this.delta,
		min:   this.min,
		max:   this.max,
	}
}

func (this *Int64) Equal(other interface{}) bool { return *this == *other.(*Int64) }
func (this *Int64) IsNumeric() bool              { return true }
func (this *Int64) IsCommutative() bool          { return true }
func (this *Int64) IsBounded() bool              { return this.min != math.MinInt64 || this.max != math.MaxInt64 }

func (this *Int64) Value() interface{} { return this.value }
func (this *Int64) Delta() interface{} { return this.delta }
func (this *Int64) DeltaSign() bool    { return this.delta >= 0 }
func (this *Int64) Min() interface{}   { return this.min }
func (this *Int64) Max() interface{}   { return this.max }

func (this *Int64) CloneDelta() interface{} { return (this.delta) }
func (this *Int64) SetValue(v interface{})  { this.value = v.(int64) }

func (this *Int64) IsDeltaApplied() bool       { return this.delta == 0 }
func (this *Int64) ResetDelta()                { this.SetDelta(common.New[int64](0)) }
func (this *Int64) SetDelta(v interface{})     { this.delta = (v.(int64)) }
func (this *Int64) SetDeltaSign(v interface{}) {}
func (this *Int64) SetMin(v interface{})       { this.min = v.(int64) }
func (this *Int64) SetMax(v interface{})       { this.max = v.(int64) }

func (this *Int64) MemSize() uint32                                            { return 5 * 8 }
func (this *Int64) TypeID() uint8                                              { return INT64 }
func (this *Int64) IsSelf(key interface{}) bool                                { return true }
func (this *Int64) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }
func (this *Int64) Reset()                                                     { this.delta = 0 }

func (this *Int64) Get() (interface{}, uint32, uint32) {
	return int64(this.value + this.delta), 1, common.IfThen(this.delta == 0, uint32(0), uint32(1))
}

func (this *Int64) Set(v interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if this.isUnderflow(int64(v.(*Int64).delta)) || this.isOverflow(int64(v.(*Int64).delta)) {
		return this, 0, 1, 0, errors.New("Error: Value out of range!!")
	}

	if (this.min > this.value+this.delta+v.(*Int64).delta) ||
		(this.max < this.value+this.delta+v.(*Int64).delta) {
		return this, 0, 1, 0, errors.New("Error: Value out of range!!")
	}

	this.delta += v.(*Int64).delta
	return this, 0, 1, 0, nil
}

func (this *Int64) isOverflow(delta int64) bool {
	flag := this.max-delta < this.value+this.delta
	return (delta >= 0 && this.delta >= 0) &&
		(this.max < delta || flag)
}

func (this *Int64) isUnderflow(delta int64) bool {
	flag := this.min-delta > this.value+this.delta
	return (delta < 0 && this.delta < 0) &&
		(this.min > delta || flag)
}

func (this *Int64) ApplyDelta(typedVals []intf.Type) (intf.Type, int, error) {
	// vec := v.([]*univalue.Univalue)
	for i, v := range typedVals {
		// v := typedVals[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Int64)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			if _, _, _, _, err := this.Set(v, nil); err != nil {
				return nil, i, err
			}
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	if this == nil {
		return nil, 0, errors.New("Error: A commutative int64 can't be nil")
	}

	this.value += this.delta
	this.delta = 0
	return this, len(typedVals), nil
}

func (this *Int64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}
