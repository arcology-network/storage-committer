package noncommutative

import (
	"bytes"
	"fmt"

	"github.com/arcology-network/common-lib/codec"
	"github.com/ethereum/go-ethereum/rlp"
)

func (this *String) Size() uint32 {
	return uint32(len(*this))
}

func (this *String) Encode() []byte {
	return codec.String(string(*this)).Encode()
}

func (this *String) EncodeToBuffer(buffer []byte) int {
	return codec.String(*this).EncodeToBuffer(buffer)
}

func (this *String) Decode(buffer []byte) interface{} {
	if len(buffer) == 0 {
		return this
	}

	*this = String(codec.String("").Decode(bytes.Clone(buffer)).(codec.String))
	return this
}

func (this *String) StorageEncode() []byte {
	buffer, _ := rlp.EncodeToBytes(*this)
	return buffer
}

func (this *String) StorageDecode(buffer []byte) interface{} {
	var v String
	rlp.DecodeBytes(buffer, &v)
	return &v
}

func (*String) Reset() {}

func (this *String) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}

func (this *String) Print() {
	fmt.Println(*this)
	fmt.Println()
}
