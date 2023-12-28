package ccurltest

import (
	"github.com/arcology-network/concurrenturl/interfaces"
	storage "github.com/arcology-network/concurrenturl/storage"
	// trie "github.com/ethereum/go-ethereum/trie"
)

const (
	ROOT_PATH   = "./tmp/filedb/"
	BACKUP_PATH = "./tmp/filedb-back/"
)

var (
	// encoder = committercommon.Codec{}.Encode
	// decoder = committercommon.Codec{}.Decode

	encoder = storage.Rlp{}.Encode
	decoder = storage.Rlp{}.Decode
)

func chooseDataStore() interfaces.Datastore {
	return storage.NewParallelEthMemDataStore() // Eth trie datastore
	// return storage.NewDataStore(nil, storage.NewCachePolicy(1000000, 1), storage.NewMemDB(), encoder, decoder)
	// return storage.NewDataStore(nil, storage.NewCachePolicy(0, 1), storage.NewMemDB(), encoder, decoder)
}
