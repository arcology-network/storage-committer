package noncommutative

import (
	"bytes"
	"fmt"

	"github.com/arcology-network/common-lib/codec"
)

func (this *String) Size(...interface{}) uint32 {
	return uint32(len(*this))
}

func (this *String) Encode(...interface{}) []byte {
	return codec.String(string(*this)).Encode()
}

func (this *String) EncodeToBuffer(buffer []byte, _ ...interface{}) int {
	return codec.String(*this).EncodeToBuffer(buffer)
}

func (this *String) Decode(buffer []byte) interface{} {
	*this = String(codec.String("").Decode(bytes.Clone(buffer)).(codec.String))
	return this
}

// func (this *String) Encode() []byte {
// 	return this.Encode()
// }

// func (this *String) DecodeCompact(bytes []byte) interface{} {
// 	return this.Decode(bytes)
// }

func (*String) Purge() {}

func (this *String) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.Encode())
}

func (this *String) Print() {
	fmt.Println(*this)
	fmt.Println()
}
