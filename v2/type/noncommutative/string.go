package noncommutative

import (
	"fmt"

	"github.com/HPISTechnologies/common-lib/codec"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
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

func (this *String) ApplyDelta(tx uint32, others []ccurlcommon.UnivalueInterface) ccurlcommon.TypeInterface {
	for _, other := range others {
		if other != nil && other.Value() != nil {
			this.Set(tx, "", other.Value().(*String), nil)
		}
	}
	return this
}

func (this *String) Composite() bool { return false }

func (this *String) Encode() []byte {
	return codec.String(string(*this)).Encode()
}

func (this *String) Decode(bytes []byte) interface{} {
	*this = String(codec.String("").Decode(bytes))
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
