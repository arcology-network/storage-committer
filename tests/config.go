package ccurltest

import (
	"github.com/arcology-network/concurrenturl/interfaces"
	storage "github.com/arcology-network/concurrenturl/storage"
)

const (
	ROOT_PATH   = "./tmp/filedb/"
	BACKUP_PATH = "./tmp/filedb-back/"
)

var (
	// encoder = storage.Codec{}.Encode
	// decoder = storage.Codec{}.Decode

	encoder = storage.Rlp{}.Encode
	decoder = storage.Rlp{}.Decode
)

func chooseDataStore() interfaces.Datastore {
	return storage.NewEthMemDataStore(false) // Eth trie datastore
	// return  cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(1000000, 1), cachedstorage.NewMemDB(), encoder, decoder)
	// return cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), cachedstorage.NewMemDB(), encoder, decoder)
}
