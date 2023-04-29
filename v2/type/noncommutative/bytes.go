package noncommutative

import (
	"bytes"

	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

//type Bytes []byte

type Bytes struct {
	placeholder bool //
	value       []byte
}

func NewBytes(v []byte) interface{} {
	b := make([]byte, len(v))
	copy(b, v)
	return &Bytes{
		placeholder: true,
		value:       b,
	}
}

func (this *Bytes) Assign(v []byte) {
	this.value = v
}

func (this *Bytes) IsSelf(key interface{}) bool { return true }
func (this *Bytes) TypeID() uint8               { return ccurlcommon.NonCommutative{}.Bytes() }

func (this *Bytes) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

// create a new path
func (this *Bytes) Clone() interface{} {
	return &Bytes{
		placeholder: true,
		value:       common.Clone(this.value),
	}
}

func (this *Bytes) Equal(other interface{}) bool {
	return bytes.Equal(this.value, other.(*Bytes).value)
}

func (this *Bytes) Get() (interface{}, uint32, uint32) {
	return this, 1, 0
}

func (this *Bytes) Value() interface{} { return this }
func (this *Bytes) Delta() interface{} {
	return this
}

func (this *Bytes) Set(value interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if value != nil && this != value { // Avoid self copy.
		this.value = make([]byte, len(value.(*Bytes).value))
		copy(this.value, value.(*Bytes).value)
	}
	return this, 0, 1, 0, nil
}

func (this *Bytes) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	vec := v.([]ccurlcommon.UnivalueInterface)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Bytes)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(v.(*Bytes), nil)
		}

		if this != nil && v == nil {
			this = nil
		}
	}

	if this == nil {
		return nil
	}
	return this
}
