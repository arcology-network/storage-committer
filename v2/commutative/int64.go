package commutative

import (
	"errors"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Int64 struct {
	value codec.Int64
	delta codec.Int64
	min   codec.Int64
	max   codec.Int64
}

func NewInt64(min, max int64) interface{} {
	if min >= max {
		return nil
	}

	return &Int64{
		min: codec.Int64(min),
		max: codec.Int64(max),
	}
}

func NewInt64Delta(delta int64) interface{}     { return &Int64{delta: codec.Int64(delta)} }
func (this *Int64) TypeID() uint8               { return INT64 }
func (this *Int64) IsSelf(key interface{}) bool { return true }

func (this *Int64) Equal(other interface{}) bool {
	return this.value == other.(*Int64).value && this.delta == other.(*Int64).delta
}

func (this *Int64) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

func (this *Int64) Clone() interface{} {
	return &Int64{
		this.value,
		this.delta,
		this.min,
		this.max,
	}
}

func (this *Int64) Get() (interface{}, uint32, uint32) {
	return int64(this.value + this.delta), 1, common.IfThen(this.delta == 0, uint32(0), uint32(1))
}

func (this *Int64) MemSize() uint32    { return 5 * 8 }
func (this *Int64) Value() interface{} { return this }
func (this *Int64) Delta() interface{} { return codec.Int64(this.delta) }

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
	flag := this.max-codec.Int64(delta) < this.value+this.delta
	return (delta >= 0 && this.delta >= 0) &&
		(this.max < codec.Int64(delta) || flag)
}

func (this *Int64) isUnderflow(delta int64) bool {
	flag := this.min-codec.Int64(delta) > this.value+this.delta
	return (delta < 0 && this.delta < 0) &&
		(this.min > codec.Int64(delta) || flag)
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
