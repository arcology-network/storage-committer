package noncommutative

import (
	"fmt"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
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

func (this *String) Decode(bytes []byte) interface{} {
	*this = String(codec.String("").Decode(common.ArrayCopy(bytes)).(codec.String))
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
