package noncommutative

import (
	"fmt"

	"github.com/arcology/common-lib/codec"
	"github.com/arcology/common-lib/common"
	ccurlcommon "github.com/arcology/concurrenturl/v2/common"
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

func (this *String) Size() uint32 {
	return uint32(len(*this))
}

func (this *String) Value() interface{} {
	return this
}

func (this *String) ToAccess() interface{} {
	return nil
}

func (this *String) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	return this, 1, 0
}

func (this *String) Peek(source interface{}) interface{} {
	return this
}

func (this *String) Delta(source interface{}) interface{} {
	return this
}

func (this *String) Set(tx uint32, path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value != nil {
		*this = String(*(value.(*String)))
	}
	return 0, 1, nil
}

func (this *String) ApplyDelta(tx uint32, v interface{}) ccurlcommon.TypeInterface {
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
			this.Set(tx, "", v.(*String), nil)
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

func (this *String) Encode() []byte {
	return codec.String(string(*this)).Encode()
}

func (this *String) EncodeToBuffer(buffer []byte) {
	codec.String(*this).EncodeToBuffer(buffer)
}

func (this *String) Decode(bytes []byte) interface{} {
	*this = String(codec.String("").Decode(common.ArrayCopy(bytes)).(codec.String))
	return this
}

func (this *String) EncodeCompact() []byte {
	return this.Encode()
}

func (this *String) DecodeCompact(bytes []byte) interface{} {
	return this.Decode(bytes)
}

func (*String) Purge() {}

func (this *String) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

func (this *String) Print() {
	fmt.Println(*this)
	fmt.Println()
}
