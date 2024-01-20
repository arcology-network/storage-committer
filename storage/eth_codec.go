package storage

import (
	"github.com/arcology-network/concurrenturl/interfaces"
)

type Rlp struct{}

func (Rlp) Encode(key string, v interface{}) []byte {
	if v == nil {
		return []byte{} // Deletion
	}
	return v.(interfaces.Type).StorageEncode(key)
}

func (Rlp) Decode(key string, buffer []byte, T any) interface{} {
	return T.(interfaces.Type).StorageDecode(key, buffer)
}
