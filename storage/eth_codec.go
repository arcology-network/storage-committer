package storage

import (
	"github.com/arcology-network/concurrenturl/interfaces"
)

type Rlp struct{}

func (Rlp) Encode(key string, v interface{}) []byte {
	if v == nil {
		return []byte{} // Deletion
	}
	return v.(interfaces.Type).StorageEncode()
}

func (Rlp) Decode(buffer []byte, T any) interface{} {
	return T.(interfaces.Type).StorageDecode(buffer)
}
