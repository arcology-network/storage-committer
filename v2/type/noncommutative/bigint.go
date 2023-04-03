package noncommutative

import (
	"math/big"

	codec "github.com/arcology-network/common-lib/codec"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Bigint big.Int

func NewBigint(v int64) interface{} {
	var value big.Int
	value.SetInt64(v)
	this := Bigint(value)
	return &this
}

func (this *Bigint) IsSelf(key interface{}) bool         { return true }
func (this *Bigint) TypeID() uint8                       { return uint8(ccurlcommon.NoncommutativeBigint) }
func (this *Bigint) This(source interface{}) interface{} { return this }
func (this *Bigint) Composite() bool                     { return false }
func (this *Bigint) Value() interface{}                  { return this }
func (this *Bigint) ToAccess() interface{}               { return nil }

func (this *Bigint) Deepcopy() interface{} {
	value := *this
	return (*Bigint)(&value)
}

func (this *Bigint) Size() uint32 {
	v := codec.Bigint(*this)
	return v.Size()
}

// create a new path
func (this *Bigint) New(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case big.Int:
		*this = Bigint(v)
	}
	return this, nil
}

func (this *Bigint) Get(path string, source interface{}) (interface{}, uint32, uint32) {
	return this, 1, 0
}

func (this *Bigint) Delta(source interface{}) interface{} {
	return this
}

func (this *Bigint) Set(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value != nil {
		*this = Bigint(*(value.(*big.Int)))
	}
	return 0, 1, nil
}

func (this *Bigint) Reset(path string, value interface{}, source interface{}) (uint32, uint32, error) {
	return this.Set(path, value, source)
}

func (this *Bigint) ApplyDelta(v interface{}) ccurlcommon.TypeInterface {
	vec := v.([]ccurlcommon.UnivalueInterface)
	for i := 0; i < len(vec); i++ {
		v := vec[i].Value()
		if this == nil && v != nil { // New value
			this = v.(*Bigint)
		}

		if this == nil && v == nil {
			this = nil
		}

		if this != nil && v != nil {
			this.Set("", v.(*Bigint), nil)
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
