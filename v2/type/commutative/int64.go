package commutative

import (
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Int64 struct {
	value int64
	delta int64
}

func NewInt64(value, delta int64) interface{} {
	return &Int64{
		value,
		delta,
	}
}

func NewInt64Delta(delta int64) interface{} {
	return &Int64{
		0,
		delta,
	}
}

func (this *Int64) TypeID() uint8               { return ccurlcommon.CommutativeInt64 }
func (this *Int64) IsSelf(key interface{}) bool { return true }

func (this *Int64) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

func (this *Int64) Deepcopy() interface{} {
	return &Int64{
		this.value,
		this.delta,
	}
}

func (this *Int64) Value() interface{} {
	return this.value
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

func (this *Int64) Latest(source interface{}) interface{} {
	return &Int64{

		this.value + this.delta,
		0,
	}
}

func (this *Int64) Delta() interface{} {
	return &Int64{

		0,
		this.delta,
	}
}

func (this *Int64) Set(v interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	this.delta += v.(*Int64).delta
	return this, 0, 1, 0, nil
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
