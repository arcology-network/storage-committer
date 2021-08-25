package noncommutative

import (
	"fmt"

	"github.com/HPISTechnologies/common-lib/codec"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/common"
)

type Int64 int64

func (this *Int64) TypeID() uint8 {
	return ccurlcommon.NoncommutativeInt64
}

func NewInt64(v int64) interface{} {
	var this Int64 = Int64(v)
	return &this
}

// create a new path
func (this *Int64) Deepcopy() interface{} {
	value := *this
	return (*Int64)(&value)
}

func (this *Int64) Value() interface{} {
	return *this
}

func (this *Int64) ToAccess() interface{} {
	return nil
}

func (this *Int64) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	return this, 1, 0
}

func (this *Int64) Transitional(source interface{}) interface{} {
	return nil
}

func (this *Int64) Set(tx uint32, path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value != nil {
		*this = Int64(*(value.(*Int64)))
	}
	return 0, 1, nil
}

func (this *Int64) Merge(tx uint32, other interface{}) {
	this.Set(tx, "", other.(ccurlcommon.TypeInterface).Value(), nil)
}

func (this *Int64) Composite() bool { return false }
func (this *Int64) Finalize()       {}

func (this *Int64) Encode() []byte {
	return codec.Int64(int64(*this)).Encode()
}

func (*Int64) Decode(bytes []byte) interface{} {
	this := Int64(codec.Int64(0).Decode(bytes))
	return &this
}

func (this *Int64) EncodeStripped() []byte {
	return this.Encode()
}

func (this *Int64) DecodeStripped(bytes []byte) interface{} {
	return this.Decode(bytes)
}

func (*Int64) Purge() {}

func (this *Int64) Print() {
	fmt.Println(*this)
	fmt.Println()
}
func (this *Int64) GobEncode() ([]byte, error) {
	return this.Encode(), nil
}
func (this *Int64) GobDecode(data []byte) error {
	myInt := this.Decode(data).(*Int64)
	*this = *myInt
	return nil
}
