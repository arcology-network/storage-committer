package ccstorage

import (
	"errors"

	common "github.com/arcology-network/common-lib/common"
	"github.com/cespare/xxhash/v2"

	slice "github.com/arcology-network/common-lib/exp/slice"
	cache "github.com/arcology-network/common-lib/storage/cache"
	commonintf "github.com/arcology-network/common-lib/storage/interface"
	policy "github.com/arcology-network/common-lib/storage/policy"
)

type DataStore struct {
	db    commonintf.PersistentStorage
	cache *cache.ReadCache[string, any]

	encoder func(string, interface{}) []byte
	decoder func(string, []byte, any) interface{}
}

// numShards uint64, isNil func(V) bool, hasher func(K) uint64, cachePolicy *policy.CachePolicy
func NewDataStore(
	cachePolicy *policy.CachePolicy,
	db commonintf.PersistentStorage,
	encoder func(string, any) []byte,
	decoder func(string, []byte, any) interface{},
) *DataStore {
	dataStore := &DataStore{
		cache: cache.NewReadCache(
			16,
			func(T any) bool {
				return T == nil
			},
			func(k string) uint64 {
				return uint64(xxhash.Sum64String(k))
			},
			cachePolicy,
		),

		db:      db,
		encoder: encoder,
		decoder: decoder,
	}

	dataStore.cache.Disable()
	return dataStore
}

// Placeholder only
func (this *DataStore) Preload(data []byte) interface{}                   { return nil }
func (this *DataStore) Cache(any) interface{}                             { return this.cache }
func (this *DataStore) Encoder(any) func(string, interface{}) []byte      { return this.encoder }
func (this *DataStore) Decoder(any) func(string, []byte, any) interface{} { return this.decoder }

func (this *DataStore) IfExists(key string) bool {
	v, _ := this.Retrive(key, nil)
	return v != nil
}

// Inject directly to the local cache.
func (this *DataStore) Inject(key string, v interface{}) error {
	this.cache.Set(key, v)
	return this.db.BatchSet([]string{key}, [][]byte{this.encoder(key, v)})
}

// Inject directly to the local cache.
func (this *DataStore) BatchInject(keys []string, values []interface{}) error {
	this.cache.BatchSet(keys, values) // update the local cache
	encoded := make([][]byte, len(keys))
	for i := 0; i < len(keys); i++ {
		encoded[i] = this.encoder(keys[i], values[i])
	}
	return this.db.BatchSet(keys, encoded)
}

func (this *DataStore) retriveFromStorage(key string, T any) (interface{}, error) {
	if this.db == nil {
		return nil, errors.New("Error: DB not found")
	}

	bytes, err := this.db.Get(key) // Get from the underlying storage
	if len(bytes) > 0 && err == nil {
		if T == nil {
			return bytes, nil
		}
		return this.decoder(key, bytes, T), nil
	}
	return nil, err
}

func (this *DataStore) Retrive(key string, T any) (interface{}, error) {
	// Read from the local cache first
	if v, _ := this.cache.Get(key); v != nil {
		return *v, nil
	}

	v, err := this.retriveFromStorage(key, T)
	if err == nil && T != nil {
		this.cache.Set(key, v) //update to the local cache and add all the missing values to the cache
	}
	return v, err
}

func (this *DataStore) BatchRetrive(keys []string, T []any) []interface{} {
	values := common.FilterFirst(this.cache.BatchGet(keys)) // From the local cache first
	if slice.Count[any, int](values, nil) == 0 {            // All found
		return values
	}

	/* Find the values missing from the local cache*/
	queryKeys, queryIdxes := make([]string, 0, len(keys)), make([]int, 0, len(keys))
	for i := 0; i < len(keys); i++ {
		if values[i] == nil {
			queryKeys = append(queryKeys, keys[i])
			queryIdxes = append(queryIdxes, i)
		}
	}

	if data, err := this.db.BatchGet(queryKeys); err == nil { // search for the values that aren't in the cache
		for i, idx := range queryIdxes {
			if data[i] != nil {
				if len(T) > 0 {
					values[idx] = this.decoder(queryKeys[i], data[i], T[i])
				} else {
					values[idx] = this.decoder(queryKeys[i], data[i], nil)
				}
			}
		}
		this.cache.BatchSet(keys, values) //update to the local cache and add all the missing values to the cache
	}
	return values
}

// func (this *DataStore) Checksum() [32]byte {
// 	return this.cache.Checksum(
// 		func(k0 string, k1 string) bool { return k0 < k1 },
// 		func(k string, v any) ([]byte, []byte) { return []byte(k), this.encoder(k, v) },
// 	)
// }

// func (this *DataStore) Query(pattern string, condition func(string, string) bool) ([]string, [][]byte, error) {
// 	return this.db.Query(pattern, condition)
// }

// Pleaseholder only
// func (this *DataStore) GetNewIndex(store intf.Datastore) interface {
// 	Add([]*univalue.Univalue)
// 	Clear()
// } {
// 	return NewCCIndexer(this, 0)
// }

// func (this *DataStore) NewWriter(blockNum uint64) interface{} {
// 	return NewAsyncWriter(this, blockNum)
// }

// func (this *DataStore) RefreshCache(blockNum uint64) (uint64, uint64) {
// 	return this.cache.Policy().Refresh(this.Cache(nil).(*mapi.ConcurrentMap[string, any]))
// }

// func (this *DataStore) CheckSum() [32]byte {
// 	k, vs := this.KVs()
// 	kData := codec.Strings(k).Flatten()
// 	vData := make([][]byte, len(vs))
// 	for i, v := range vs {
// 		vData[i] = this.encoder(k[i], v)
// 	}
// 	vData = append(vData, kData)
// 	return sha256.Sum256(codec.Byteset(vData).Flatten())
// }

// func (this *DataStore) KVs() ([]string, []interface{}) {
// 	return this.cache.KVs()
// }
