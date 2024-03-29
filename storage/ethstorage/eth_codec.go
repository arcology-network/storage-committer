package ethstorage

import (
	"github.com/arcology-network/storage-committer/interfaces"
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
