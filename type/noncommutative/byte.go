package noncommutative

import (
	"fmt"

	ccurlcommon "github.com/arcology/concurrenturl/common"
)

type Bytes []byte

func (this *Bytes) TypeID() uint8 {
	return ccurlcommon.NoncommutativeBytes
}

func NewBytes(v []byte) interface{} {
	var this Bytes = Bytes(v)
	return &this
}

// create a new path
func (this *Bytes) Deepcopy() interface{} {
	value := make([]byte, len(*this))
	copy(value, *this)
	return (*Bytes)(&value)
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

func (this *Bytes) Transitional(source interface{}) interface{} {
	return nil
}

func (this *Bytes) Set(tx uint32, path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value != nil && this != value { // Avoid self copy.
		*this = make([]byte, len(Bytes(*(value.(*Bytes)))))
		copy(*this, Bytes(*(value.(*Bytes))))
	}
	return 0, 1, nil
}

func (this *Bytes) Merge(tx uint32, other interface{}) {
	this.Set(tx, "", other.(ccurlcommon.TypeInterface).Value(), nil)
}

func (this *Bytes) Composite() bool { return false }
func (this *Bytes) Finalize()       {}

func (this *Bytes) Encode() []byte {
	return *this
}

func (*Bytes) Decode(bytes []byte) interface{} {
	var this Bytes = Bytes(bytes)
	return &this
}

func (this *Bytes) GobEncode() ([]byte, error) {
	return this.Encode(), nil
}
func (this *Bytes) GobDecode(data []byte) error {
	mybytes := this.Decode(data).(*Bytes)
	*this = *mybytes
	return nil
}

func (this *Bytes) EncodeStripped() []byte {
	return this.Encode()
}

func (this *Bytes) DecodeStripped(bytes []byte) interface{} {
	return this.Decode(bytes)
}

func (*Bytes) Purge() {}

func (this *Bytes) Print() {
	fmt.Println(*this)
	fmt.Println()
}
