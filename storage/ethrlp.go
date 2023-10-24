package ccdb

import (
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/evm/rlp"
)

type Rlp struct{}

func (Rlp) Encode(v interface{}) []byte {
	return common.Return1st(rlp.EncodeToBytes(v))
}

func (Rlp) Decode(data []byte, T any) interface{} {
	err := rlp.DecodeBytes(data, T)
	if err != nil {
		return nil
	}
	return T
}

// func (this Rlp) Encoder(v interface{}) func(interface{}) []byte {
// 	return this.Encode
// }

// func (this Rlp) Decoder(v interface{}) func([]byte) (interface{}, error) {
// 	this.V = v
// 	return this.Decode
// }
