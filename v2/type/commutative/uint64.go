package commutative

import (
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Uint64 struct {
	finalized bool
	value     uint64
	delta     uint64
}

func NewUint64(value uint64, delta uint64) interface{} {
	return &Uint64{
		false,
		value,
		delta,
	}
}

func (this *Uint64) TypeID() uint8               { return ccurlcommon.CommutativeUint64 }
func (this *Uint64) IsSelf(key interface{}) bool { return true }

func (this *Uint64) Deepcopy() interface{} {
	return &Uint64{
		this.finalized,
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

func (this *Uint64) Get(path string, source interface{}) (interface{}, uint32, uint32) {
	if this.delta == 0 {
		return this, 1, 0
	}

	this.finalized = true
	return &Uint64{
		finalized: true,
		value:     this.value + this.delta,
		delta:     0,
	}, 1, 1
}

func (this *Uint64) This(source interface{}) interface{} {
	return &Uint64{
		this.finalized,
		this.value + this.delta,
		0,
	}
}

func (this *Uint64) Delta() interface{} {
	return &Uint64{
		this.finalized,
		0,
		this.delta,
	}
}

func (this *Uint64) Set(path string, v interface{}, source interface{}) (uint32, uint32, error) {
	this.delta += v.(*Uint64).delta
	return 0, 1, nil
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
			this.Set("", v.(*Uint64), nil)
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

func (this *Uint64) ConcurrentWritable() bool { return !this.finalized }

func (this *Uint64) Purge() {
	this.finalized = false
	this.delta = 0
}

func (this *Uint64) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}
