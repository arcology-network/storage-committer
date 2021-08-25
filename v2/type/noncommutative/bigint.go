package noncommutative

import (
	"fmt"
	"math/big"

	codec "github.com/arcology/common-lib/codec"
	ccurlcommon "github.com/arcology/concurrenturl/v2/common"
)

type Bigint big.Int

func NewBigint(v int64) interface{} {
	var value big.Int
	value.SetInt64(v)
	this := Bigint(value)
	return &this
}

func (this *Bigint) TypeID() uint8 { return uint8(ccurlcommon.NoncommutativeBigint) }

func (this *Bigint) Deepcopy() interface{} {
	value := *this
	return (*Bigint)(&value)
}

func (this *Bigint) Value() interface{} {
	return this
}

func (this *Bigint) ToAccess() interface{} {
	return nil
}

// create a new path
func (this *Bigint) New(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case big.Int:
		*this = Bigint(v)
	}
	return this, nil
}

func (this *Bigint) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	return this, 1, 0
}

func (this *Bigint) Delta(source interface{}) interface{} {
	return this
}

func (this *Bigint) Peek(source interface{}) interface{} {
	return this
}

func (this *Bigint) Set(tx uint32, path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value != nil {
		*this = Bigint(*(value.(*big.Int)))
	}
	return 0, 1, nil
}

func (this *Bigint) ApplyDelta(tx uint32, others []ccurlcommon.UnivalueInterface) ccurlcommon.TypeInterface {
	for _, other := range others {
		if other != nil && other.Value() != nil {
			this.Set(tx, "", other.Value().(*Bigint), nil)
		}
	}
	return this
}

func (this *Bigint) Composite() bool { return false }

func (this *Bigint) Encode() []byte {
	v := codec.Bigint(*this)
	return v.Encode()
}

func (*Bigint) Decode(bytes []byte) interface{} {
	this := Bigint(*(&codec.Bigint{}).Decode(bytes))
	return &this
}

func (this *Bigint) EncodeCompact() []byte {
	return this.Encode()
}

func (this *Bigint) DecodeCompact(bytes []byte) interface{} {
	return this.Decode(bytes)
}

func (*Bigint) Purge() {}

func (this *Bigint) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

func (this *Bigint) Print() {
	fmt.Println(*this)
	fmt.Println()
}
