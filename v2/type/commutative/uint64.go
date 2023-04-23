package commutative

import (
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Uint64 struct {
	value uint64
	delta uint64
}

func NewUint64(value uint64, delta uint64) interface{} {
	return &Uint64{
		value,
		delta,
	}
}

func (this *Uint64) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

func (this *Uint64) TypeID() uint8               { return ccurlcommon.CommutativeUint64 }
func (this *Uint64) IsSelf(key interface{}) bool { return true }

func (this *Uint64) Deepcopy() interface{} {
	return &Uint64{
		this.value,
		this.delta,
	}
}

func (this *Uint64) Value() interface{} {
	return this.value
}

func (this *Uint64) ToAccess() interface{} {
	return this
}

func (this *Uint64) Get(source interface{}) (interface{}, uint32, uint32) {
	if this.delta == 0 {
		return this, 1, 0
	}

	return &Uint64{
		value: this.value + this.delta,
		delta: 0,
	}, 1, 1
}

func (this *Uint64) Latest() interface{} {
	return &Uint64{
		this.value + this.delta,
		0,
	}
}

func (this *Uint64) Delta() interface{} {
	return &Uint64{
		0,
		this.delta,
	}
}

func (this *Uint64) Set(v interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	this.delta += v.(*Uint64).delta
	return this, 0, 1, 0, nil
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

	this.value += this.delta
	this.delta = 0
	return this
}

func (this *Uint64) Purge() {
	this.delta = 0
}

func (this *Uint64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}
