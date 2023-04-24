package noncommutative

import (
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

//type Bytes []byte

type Bytes struct {
	placeholder bool //
	data        []byte
}

func NewBytes(v []byte) interface{} {
	b := make([]byte, len(v))
	copy(b, v)
	return &Bytes{
		placeholder: true,
		data:        b,
	}
}

func (this *Bytes) Assign(v []byte) {
	this.data = v
}

func (this *Bytes) IsSelf(key interface{}) bool { return true }
func (this *Bytes) TypeID() uint8               { return ccurlcommon.NoncommutativeBytes }

func (this *Bytes) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

// create a new path
func (this *Bytes) Deepcopy() interface{} {
	return &Bytes{
		placeholder: true,
		data:        common.DeepCopy(this.data),
	}
}

func (this *Bytes) Data() []byte {
	return this.data
}

func (this *Bytes) Value() interface{} {
	return this
}

func (this *Bytes) ToAccess() interface{} {
	return nil
}

func (this *Bytes) Get(source interface{}) (interface{}, uint32, uint32) {
	return this, 1, 0
}

// func (this *Bytes) Latest() interface{} {
// 	return this
// }

func (this *Bytes) Delta() interface{} {
	return this
}

func (this *Bytes) Set(value interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if value != nil && this != value { // Avoid self copy.
		this.data = make([]byte, len(value.(*Bytes).data))
		copy(this.data, value.(*Bytes).data)
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
