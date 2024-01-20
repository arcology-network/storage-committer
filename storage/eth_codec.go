package storage

import (
	"strings"

	"github.com/arcology-network/concurrenturl/interfaces"
)

type Rlp struct{}

func (Rlp) Encode(key string, v interface{}) []byte {
	if v == nil {
		return []byte{} // Deletion
	}

	flag := strings.Contains(key, "/native/")
	return v.(interfaces.Type).StorageEncode(flag)
}

func (Rlp) Decode(isNative bool, buffer []byte, T any) interface{} {
	return T.(interfaces.Type).StorageDecode(isNative, buffer)
}
