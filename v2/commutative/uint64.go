package commutative

import (
	"errors"

	codec "github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Selector []bool

type Uint64 struct {
	value *codec.Uint64
	delta *codec.Uint64
	min   *codec.Uint64
	max   *codec.Uint64
}

func NewUint64(min, max uint64) interface{} {
	if min >= max {
		return nil
	}

	value := codec.Uint64(0)
	delta := codec.Uint64(0)
	minV := codec.Uint64(min)
	maxV := codec.Uint64(max)
	return &Uint64{
		&value,
		&delta,
		&minV,
		&maxV,
	}
}

func NewUint64Delta(delta uint64) interface{} {
	deltaV := codec.Uint64(delta)
	return &Uint64{delta: &deltaV}
}

func (this *Uint64) New(value, delta, sign, min, max interface{}) interface{} {
	return &Uint64{
		common.IfThenDo1st(value != nil, func() *codec.Uint64 { return value.(*codec.Uint64) }, nil),
		common.IfThenDo1st(delta != nil, func() *codec.Uint64 { return delta.(*codec.Uint64) }, nil),
		common.IfThenDo1st(min != nil, func() *codec.Uint64 { return min.(*codec.Uint64) }, nil),
		common.IfThenDo1st(max != nil, func() *codec.Uint64 { return max.(*codec.Uint64) }, nil),
	}
}

func (this *Uint64) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) { return v, 0, 1, 0 }

func (this *Uint64) Equal(other interface{}) bool {
	return *this.value == *other.(*Uint64).value &&
		*this.delta == *other.(*Uint64).delta &&
		*this.min == *other.(*Uint64).min &&
		*this.max == *other.(*Uint64).max
}

func (this *Uint64) MemSize() uint32 { return 5 * 8 }

func (this *Uint64) Value() interface{} { return this.value }
func (this *Uint64) Delta() interface{} { return this.delta }
func (this *Uint64) Sign() interface{}  { return true }
func (this *Uint64) Min() interface{}   { return this.min }
func (this *Uint64) Max() interface{}   { return this.max }

func (this *Uint64) TypeID() uint8               { return UINT64 }
func (this *Uint64) IsSelf(key interface{}) bool { return true }

func (this *Uint64) Clone() interface{} {
	return &Uint64{
		value: this.value,
		delta: this.delta,
		min:   this.min,
		max:   this.max,
	}
}

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

func (this *Uint64) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	vec := v.([]ccurlcommon.UnivalueInterface)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Uint64)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(v.(*Uint64), nil)
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	if this == nil {
		return nil
	}

	*this.value += *this.delta
	*this.delta = 0
	return this
}

func (this *Uint64) Purge() {
	*this.delta = 0
}

func (this *Uint64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}
