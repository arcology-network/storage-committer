package noncommutative

import (
	"fmt"

	codec "github.com/arcology-network/common-lib/codec"
)

func (this *Bigint) Encode(processors ...func(interface{}) interface{}) []byte {
	v := codec.Bigint(*this)
	return v.Encode()
}

func (this *Bigint) EncodeToBuffer(buffer []byte, processors ...func(interface{}) interface{}) int {
	v := codec.Bigint(*this)
	return v.EncodeToBuffer(buffer)
}

func (this *Bigint) Decode(bytes []byte) interface{} {
	this = (*Bigint)((&codec.Bigint{}).Decode(bytes).(*codec.Bigint))
	return this
}

func (this *Bigint) EncodeCompact() []byte {
	return this.Encode()
}

func (this *Bigint) DecodeCompact(bytes []byte) interface{} {
	return this.Decode(bytes)
}

func (this *Bigint) Purge() {}

func (this *Bigint) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
}

func (this *Bigint) Print() {
	fmt.Println(*this)
	fmt.Println()
}
