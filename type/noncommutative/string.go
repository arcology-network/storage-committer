package noncommutative

import (
	"fmt"

	"github.com/HPISTechnologies/common-lib/codec"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/common"
)

type String string

func (this *String) TypeID() uint8 {
	return uint8(ccurlcommon.NoncommutativeString)
}

func NewString(v string) interface{} {
	var this String = String(v)
	return &this
}

func (this *String) Deepcopy() interface{} {
	value := *this
	return (*String)(&value)
}

func (this *String) Value() interface{} {
	return *this
}

func (this *String) ToAccess() interface{} {
	return nil
}

func (this *String) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	return this, 1, 0
}

func (this *String) Transitional(source interface{}) interface{} {
	return nil
}

func (this *String) Set(tx uint32, path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value != nil {
		*this = String(*(value.(*String)))
	}
	return 0, 1, nil
}

func (this *String) Merge(tx uint32, other interface{}) {
	this.Set(tx, "", other.(ccurlcommon.TypeInterface).Value(), nil)
}

func (this *String) Composite() bool { return false }
func (this *String) Finalize()       {}

func (this *String) Encode() []byte {
	return codec.String(string(*this)).Encode()
}

func (this *String) Decode(bytes []byte) interface{} {
	*this = String(codec.String("").Decode(bytes))
	return this
}

func (this *String) EncodeStripped() []byte {
	return this.Encode()
}

func (this *String) DecodeStripped(bytes []byte) interface{} {
	return this.Decode(bytes)
}

func (*String) Purge() {}

func (this *String) Print() {
	fmt.Println(*this)
	fmt.Println()
}
func (this *String) GobEncode() ([]byte, error) {
	return this.Encode(), nil
}
func (this *String) GobDecode(data []byte) error {
	myString := this.Decode(data).(*String)
	*this = *myString
	return nil
}
