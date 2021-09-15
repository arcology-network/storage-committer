package noncommutative

import (
	"fmt"

	"github.com/arcology-network/common-lib/codec"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	"github.com/elliotchance/orderedmap"
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

func (this *String) ApplyDelta(tx uint32, v interface{}) ccurlcommon.TypeInterface {
	for iter := v.(*orderedmap.Element); iter != nil; iter = iter.Next() {
		if iter.Value == nil {
			continue
		}

		if v := iter.Value.(ccurlcommon.UnivalueInterface).Value(); v != nil {
			this.Set(tx, "", v.(*String), nil)
		} else {
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
