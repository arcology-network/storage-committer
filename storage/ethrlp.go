package ccdb

import (
	"github.com/arcology-network/concurrenturl/interfaces"
)

type Rlp struct{}

func (this Rlp) Encode(key string, v interface{}) []byte {
	if v == nil {
		return []byte{} // Deletion
	}

	return v.(interfaces.Type).StorageEncode()
}

func (Rlp) Decode(buffer []byte, T any) interface{} {
	return T.(interfaces.Type).StorageDecode(buffer)
}

// func (this Rlp) Encoder(v interface{}) func(interface{}) []byte {
// 	return this.Encode
// }

// func (this Rlp) Decoder(v interface{}) func([]byte) (interface{}, error) {
// 	this.V = v
// 	return this.Decode
// }
