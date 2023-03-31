package commutative

import (
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Int64 struct {
	finalized bool
	value     int64
	delta     int64
}

func NewInt64(value int64, delta int64) interface{} {
	return &Int64{
		false,
		value,
		delta,
	}
}

func (this *Int64) TypeID() uint8 { return ccurlcommon.CommutativeInt64 }

func (this *Int64) Deepcopy() interface{} {
	return &Int64{
		this.finalized,
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

func (this *Int64) Get(path string, source interface{}) (interface{}, uint32, uint32) {
	if this.delta == 0 {
		return this, 1, 0
	}

	this.finalized = true
	return &Int64{
		finalized: true,
		value:     this.value + this.delta,
		delta:     0,
	}, 1, 1
}

func (this *Int64) This(source interface{}) interface{} {
	return &Int64{
		this.finalized,
		this.value + this.delta,
		0,
	}
}

func (this *Int64) Delta(source interface{}) interface{} {
	return &Int64{
		this.finalized,
		0,
		this.delta,
	}
}

func (this *Int64) Set(path string, v interface{}, source interface{}) (uint32, uint32, error) {
	this.delta += v.(*Int64).delta
	return 0, 1, nil
}

func (this *Int64) Reset(path string, v interface{}, source interface{}) (uint32, uint32, error) {
	this.value = 0
	this.delta = v.(int64) // This is by design
	this.finalized = true
	return 0, 1, nil
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
			this.Set("", v.(*Int64), nil)
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

func (this *Int64) Composite() bool { return !this.finalized }

func (this *Int64) Purge() {
	this.finalized = false
	this.delta = 0
}

func (this *Int64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}
