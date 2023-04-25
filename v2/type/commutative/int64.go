package commutative

import (
	"errors"

	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Int64 struct {
	value int64
	delta int64
	min   int64
	max   int64
}

func NewInt64(min, max int64) interface{} {
	if min >= max {
		return nil
	}

	return &Int64{
		min: min,
		max: max,
	}
}

func NewInt64Delta(delta int64) interface{}     { return &Int64{delta: delta} }
func (this *Int64) TypeID() uint8               { return ccurlcommon.CommutativeInt64 }
func (this *Int64) IsSelf(key interface{}) bool { return true }

func (this *Int64) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

func (this *Int64) Deepcopy() interface{} {
	return &Int64{
		this.value,
		this.delta,
		this.min,
		this.max,
	}
}

func (this *Int64) ToAccess() interface{} {
	return this
}

func (this *Int64) Get(source interface{}) (interface{}, uint32, uint32) {
	if this.delta == 0 {
		return this, 1, 0
	}

	return &Int64{
		value: this.value + this.delta,
		delta: 0,
	}, 1, 1
}

func (this *Int64) Value() interface{} {
	return this.value
}

func (this *Int64) Delta() interface{} {
	return &Int64{
		0,
		this.delta,
		this.min,
		this.max,
	}
}

func (this *Int64) Set(v interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if this.isUnderflow(v.(*Int64).delta) || this.isOverflow(v.(*Int64).delta) {
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

func (this *Int64) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	vec := v.([]ccurlcommon.UnivalueInterface)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Int64)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(v.(*Int64), nil)
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	if this == nil {
		return nil
	}

	this.value += this.delta
	this.delta = 0
	return this
}

func (this *Int64) Purge() {
	this.delta = 0
}

func (this *Int64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}
