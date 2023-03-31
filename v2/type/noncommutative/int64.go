package noncommutative

import (
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Int64 int64

func NewInt64(v int64) interface{} {
	var this Int64 = Int64(v)
	return &this
}

func (this *Int64) TypeID() uint8 { return ccurlcommon.NoncommutativeInt64 }

// create a new path
func (this *Int64) Deepcopy() interface{} {
	value := *this
	return (*Int64)(&value)
}

func (this *Int64) Value() interface{} {
	return this
}

func (this *Int64) ToAccess() interface{} {
	return nil
}

func (this *Int64) Get(path string, source interface{}) (interface{}, uint32, uint32) {
	return this, 1, 0
}

func (this *Int64) This(source interface{}) interface{} {
	return this
}

func (this *Int64) Delta(source interface{}) interface{} {
	return this
}

func (this *Int64) Set(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value != nil {
		*this = Int64(*(value.(*Int64)))
	}
	return 0, 1, nil
}

func (this *Int64) Reset(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	return this.Set(path, value, source)
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
	return this
}

func (this *Int64) Composite() bool { return false }
