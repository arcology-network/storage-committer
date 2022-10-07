package noncommutative

import (
	"fmt"

	codec "github.com/HPISTechnologies/common-lib/codec"
	"github.com/HPISTechnologies/common-lib/common"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
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

func (this *Bytes) HeaderSize() uint32 {
	return 3 * codec.UINT32_LEN
}

func (this *Bytes) Size() uint32 {
	return this.HeaderSize() + uint32(1+len(this.data))
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
			this.Set(tx, "", v.(*Bytes), nil)
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

func (this *Bytes) Composite() bool { return false }

func (this *Bytes) Encode() []byte {
	byteset := [][]byte{
		codec.Bool(this.placeholder).Encode(),
		this.data,
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Bytes) EncodeToBuffer(buffer []byte) {
	codec.Encoder{}.ToBuffer(
		buffer,
		[]interface{}{
			codec.Bool(this.placeholder),
			codec.Bytes(this.data),
		},
	)
}

func (*Bytes) Decode(bytes []byte) interface{} {
	fields := codec.Byteset{}.Decode(bytes).(codec.Byteset)
	return &Bytes{
		placeholder: bool(codec.Bool(true).Decode(fields[0]).(codec.Bool)),
		data:        common.ArrayCopy(fields[1]),
	}
}

func (this *Bytes) EncodeCompact() []byte {
	return this.Encode()
}

func (this *Bytes) DecodeCompact(bytes []byte) interface{} {
	return this.Decode(bytes)
}

func (this *Bytes) Purge() {}

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
