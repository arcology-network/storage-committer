package noncommutative

import (
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	"github.com/elliotchance/orderedmap"
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

func (this *Bytes) TypeID() uint8 { return ccurlcommon.NoncommutativeBytes }

// create a new path
func (this *Bytes) Deepcopy() interface{} {
	value := make([]byte, len(this.data))
	copy(value, this.data)
	return &Bytes{
		placeholder: true,
		data:        value,
	}
}

func (this *Bytes) Value() interface{} {
	return this
}

func (this *Bytes) ToAccess() interface{} {
	return nil
}

func (this *Bytes) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	return this, 1, 0
}

func (this *Bytes) Peek(source interface{}) interface{} {
	return this
}

func (this *Bytes) Delta(source interface{}) interface{} {
	return this
}

func (this *Bytes) Set(tx uint32, path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value != nil && this != value { // Avoid self copy.
		this.data = make([]byte, len(value.(*Bytes).data))
		copy(this.data, value.(*Bytes).data)
	}
	return 0, 1, nil
}

func (this *Bytes) ApplyDelta(tx uint32, v interface{}) ccurlcommon.TypeInterface {
	for iter := v.(*orderedmap.Element); iter != nil; iter = iter.Next() {
		if iter.Value == nil {
			continue
		}

		if v := iter.Value.(ccurlcommon.UnivalueInterface).Value(); v != nil {
			this.Set(tx, "", v.(*Bytes), nil)
		} else {
			this = nil
		}
	}

	if this == nil {
		return nil
	}
	return this
}

func (this *Bytes) Composite() bool { return false }

func (this *Bytes) Encode() []byte {
	byteset := [][]byte{
		codec.Bool(this.placeholder).Encode(),
		this.data,
	}
	return codec.Byteset(byteset).Encode()
}

func (*Bytes) Decode(bytes []byte) interface{} {
	fields := codec.Byteset{}.Decode(bytes)
	return &Bytes{
		placeholder: bool(codec.Bool(true).Decode(fields[0])),
		data:        fields[1],
	}
}

func (this *Bytes) EncodeCompact() []byte {
	return this.Encode()
}

func (this *Bytes) DecodeCompact(bytes []byte) interface{} {
	return this.Decode(bytes)
}

func (*Bytes) Purge() {}

func (this *Bytes) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

func (this *Bytes) Print() {
	fmt.Println(*this)
	fmt.Println()
}

func (this *Bytes) Data() []byte {
	return this.data
}
