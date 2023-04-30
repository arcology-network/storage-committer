package noncommutative

import (
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type String string

func NewString(v string) interface{} {
	var this String = String(v)
	return &this
}

func (this *String) IsSelf(key interface{}) bool { return true }
func (this *String) TypeID() uint8               { return STRING }

func (this *String) MemSize() uint32              { return uint32(len(*this)) }
func (this *String) Value() interface{}           { return this }
func (this *String) Delta() interface{}           { return this }
func (this *String) Equal(other interface{}) bool { return *this == *(other.(*String)) }

func (this *String) CopyTo(v interface{}) (interface{}, uint32, uint32, uint32) {
	return v, 0, 1, 0
}

func (this *String) Clone() interface{} {
	value := *this
	return (*String)(&value)
}

func (this *String) Get() (interface{}, uint32, uint32) {
	return this, 1, 0
}

func (this *String) Set(value interface{}, source interface{}) (interface{}, uint32, uint32, uint32, error) {
	if value != nil {
		*this = String(*(value.(*String)))
	}
	return this, 0, 1, 0, nil
}

func (this *String) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	vec := v.([]ccurlcommon.UnivalueInterface)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*String)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set(v.(*String), nil)
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
