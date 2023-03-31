package noncommutative

import (
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type String string

func NewString(v string) interface{} {
	var this String = String(v)
	return &this
}

func (this *String) TypeID() uint8 { return uint8(ccurlcommon.NoncommutativeString) }

func (this *String) Deepcopy() interface{} {
	value := *this
	return (*String)(&value)
}

func (this *String) Value() interface{} {
	return this
}

func (this *String) ToAccess() interface{} {
	return nil
}

func (this *String) Get(path string, source interface{}) (interface{}, uint32, uint32) {
	return this, 1, 0
}

func (this *String) This(source interface{}) interface{} {
	return this
}

func (this *String) Delta(source interface{}) interface{} {
	return this
}

func (this *String) Set(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value != nil {
		*this = String(*(value.(*String)))
	}
	return 0, 1, nil
}

func (this *String) Reset(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	return this.Set(path, value, source)
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
			this.Set("", v.(*String), nil)
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

func (this *String) Composite() bool { return false }
