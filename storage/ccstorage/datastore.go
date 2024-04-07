package ccstorage

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	addrcompressor "github.com/arcology-network/common-lib/addrcompressor"
	codec "github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	intf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/univalue"
	"github.com/cespare/xxhash/v2"

	// mapi "github.com/arcology-network/common-lib/container/map"
	mapi "github.com/arcology-network/common-lib/exp/map"
	slice "github.com/arcology-network/common-lib/exp/slice"
	cache "github.com/arcology-network/common-lib/storage/cache"
	commonintf "github.com/arcology-network/common-lib/storage/interface"
	policy "github.com/arcology-network/common-lib/storage/policy"
)

type DataStore struct {
	db      commonintf.PersistentStorage
	cccache *cache.ReadCache[string, any]

	keyCompressor *addrcompressor.CompressionLut

	queue       chan *CCIndexer
	commitQueue chan *CCIndexer
	encoder     func(string, interface{}) []byte
	decoder     func(string, []byte, any) interface{}
}

// numShards uint64, isNil func(V) bool, hasher func(K) uint64, cachePolicy *policy.CachePolicy
func NewDataStore(
	keyCompressor *addrcompressor.CompressionLut,
	cachePolicy *policy.CachePolicy,
	db commonintf.PersistentStorage,
	encoder func(string, any) []byte,
	decoder func(string, []byte, any) interface{},
) *DataStore {
	dataStore := &DataStore{
		cccache: cache.NewReadCache(
			16,
			func(T any) bool {
				return T == nil
			},
			func(k string) uint64 {
				return uint64(xxhash.Sum64String(k))
			},
			cachePolicy,
		),
		queue:         make(chan *CCIndexer, 64),
		commitQueue:   make(chan *CCIndexer, 64),
		keyCompressor: keyCompressor,

		db:      db,
		encoder: encoder,
		decoder: decoder,
	}

	return dataStore
}

// Pleaseholder only
func (this *DataStore) GetNewIndex(store intf.Datastore) interface {
	Add([]*univalue.Univalue)
	Clear()
} {
	return NewIndexer(store)
}

// Pleaseholder only
func (this *DataStore) Preload(data []byte) interface{}                   { return nil }
func (this *DataStore) Cache(any) interface{}                             { return this.cccache }
func (this *DataStore) Encoder(any) func(string, interface{}) []byte      { return this.encoder }
func (this *DataStore) Decoder(any) func(string, []byte, any) interface{} { return this.decoder }

func (this *DataStore) Size() uint64 { return uint64(this.cccache.Length()) }

func (this *DataStore) Checksum() [32]byte {
	return this.cccache.Checksum(
		func(k0 string, k1 string) bool { return k0 < k1 },
		func(k string, v any) ([]byte, []byte) { return []byte(k), this.encoder(k, v) },
	)
}

func (this *DataStore) Query(pattern string, condition func(string, string) bool) ([]string, [][]byte, error) {
	return this.db.Query(pattern, condition)
}

func (this *DataStore) IfExists(key string) bool {
	v, _ := this.Retrive(key, nil)
	return v != nil
}

// Inject directly to the local cache.
func (this *DataStore) Inject(key string, v interface{}) error {
	if this.keyCompressor != nil {
		key = this.keyCompressor.CompressOnTemp([]string{key})[0]
		this.keyCompressor.Commit()
	}

	this.cccache.Set(key, v)
	return this.db.BatchSet([]string{key}, [][]byte{this.encoder(key, v)})
}

// Inject directly to the local cache.
func (this *DataStore) BatchInject(keys []string, values []interface{}) error {
	if this.keyCompressor != nil {
		this.keyCompressor.CompressOnTemp(keys)
		this.keyCompressor.Commit()
	}

	// this.batchAddToCache(this.GetPartitions(keys), keys, values)
	this.cccache.BatchSet(keys, values) // update the local cache
	encoded := make([][]byte, len(keys))
	for i := 0; i < len(keys); i++ {
		encoded[i] = this.encoder(keys[i], values[i])
	}
	return this.db.BatchSet(keys, encoded)
}

func (this *DataStore) RetriveFromStorage(key string, T any) (interface{}, error) {
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
	if this.keyCompressor != nil {
		key = this.keyCompressor.TryCompress(key) // Convert the key
	}

	// Read from the local cache first
	if v, _ := this.cccache.Get(key); v != nil {
		return *v, nil
	}

	v, err := this.RetriveFromStorage(key, T)
	if err == nil && T != nil {
		this.cccache.Set(key, v) //update to the local cache and add all the missing values to the cache
	}
	return v, err
}

func (this *DataStore) BatchRetrive(keys []string, T []any) []interface{} {
	if this.keyCompressor != nil {
		keys = this.keyCompressor.TryBatchCompress(keys)
	}

	values := common.FilterFirst(this.cccache.BatchGet(keys)) // From the local cache first
	if slice.Count[any, int](values, nil) == 0 {              // All found
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
		this.cccache.BatchSet(keys, values) //update to the local cache and add all the missing values to the cache
	}
	return values
}

func (this *DataStore) Precommit(arg ...interface{}) [32]byte {
	arg[0].(*CCIndexer).Get() // To remove some empty transitions
	return [32]byte{}
}

// Commit the changes to the local cache and the persistent storage
func (this *DataStore) Commit(_ uint64) error {
	for len(this.commitQueue) != 0 {
		fmt.Println("Waiting for the commit job queue to be emptied")
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (this *DataStore) CommitV2(idx *CCIndexer) error {
	var err error
	if this.keyCompressor != nil {
		common.ParallelExecute(
			func() { err = this.db.BatchSet(idx.keyBuffer, idx.encodedBuffer) }, // Write data back
			func() { this.keyCompressor.Commit() })

	} else {
		err = this.db.BatchSet(idx.keyBuffer, idx.encodedBuffer)
	}

	this.cccache.BatchSet(idx.keyBuffer, idx.valueBuffer) // update the local cache
	return err
}

func (this *DataStore) RefreshCache(blockNum uint64) (uint64, uint64) {
	return this.CachePolicy().Refresh(this.Cache(nil).(*mapi.ConcurrentMap[string, any]))
}

func (this *DataStore) Print() {
	this.cccache.Print()
}

func (this *DataStore) CheckSum() [32]byte {
	k, vs := this.KVs()
	kData := codec.Strings(k).Flatten()
	vData := make([][]byte, len(vs))
	for i, v := range vs {
		vData[i] = this.encoder(k[i], v)
	}
	vData = append(vData, kData)
	return sha256.Sum256(codec.Byteset(vData).Flatten())
}

func (this *DataStore) KVs() ([]string, []interface{}) {
	return this.cccache.KVs()
}

func (this *DataStore) CachePolicy() *policy.CachePolicy {
	return this.cccache.Policy()
}
