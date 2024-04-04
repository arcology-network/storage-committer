package ccstorage

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
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
	db   commonintf.PersistentStorage
	lock sync.RWMutex

	keyCompressor *addrcompressor.CompressionLut
	cccache       *cache.ReadCache[string, any]
	queue         chan *CCIndexer
	encoder       func(string, interface{}) []byte
	decoder       func(string, []byte, any) interface{}

	partitionIDs []uint64

	keyBuffer     []string
	valueBuffer   []interface{}
	encodedBuffer [][]byte //The encoded buffer contains the encoded values

	commitLock sync.RWMutex
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
		partitionIDs: make([]uint64, 0, 65536),
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
	this.lock.RLock()
	defer this.lock.RUnlock()
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

	this.addToCache(key, v)
	return this.batchWritePersistentStorage([]string{key}, [][]byte{this.encoder(key, v)})
}

// Inject directly to the local cache.
func (this *DataStore) BatchInject(keys []string, values []interface{}) error {
	if this.keyCompressor != nil {
		this.keyCompressor.CompressOnTemp(keys)
		this.keyCompressor.Commit()
	}

	this.batchAddToCache(this.GetPartitions(keys), keys, values)
	encoded := make([][]byte, len(keys))
	for i := 0; i < len(keys); i++ {
		encoded[i] = this.encoder(keys[i], values[i])
	}
	return this.batchWritePersistentStorage(keys, encoded)
}

func (this *DataStore) RetriveFromStorage(key string, T any) (interface{}, error) {
	if this.db == nil {
		return nil, errors.New("Error: DB not found")
	}

	this.lock.RLock()
	bytes, err := this.db.Get(key) // Get from the underlying storage
	this.lock.RUnlock()

	if len(bytes) > 0 && err == nil {
		if T == nil {
			return bytes, nil
		}
		return this.decoder(key, bytes, T), nil
	}
	return nil, err
}

func (this *DataStore) batchFetchPersistentStorage(keys []string) ([][]byte, error) {
	if this.db == nil {
		return nil, errors.New("Error: DB not found")
	}

	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.db.BatchGet(keys) // Get from the cache
}

func (this *DataStore) batchWritePersistentStorage(keys []string, encodedValues [][]byte) error {
	if this.db == nil {
		return errors.New("Error: DB not found")
	}

	this.lock.Lock()
	defer this.lock.Unlock()
	return this.db.BatchSet(keys, encodedValues)
}

func (this *DataStore) addToCache(key string, value interface{}) {
	this.cccache.Set(key, value)
}

func (this *DataStore) batchAddToCache(ids []uint64, keys []string, values []interface{}) {
	this.cccache.BatchGet(keys, values)
}

func (this *DataStore) Buffers() ([]string, []interface{}, [][]byte) {
	return this.keyBuffer, this.valueBuffer, this.encodedBuffer
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
		this.addToCache(key, v) //update to the local cache and add all the missing values to the cache
	}
	return v, err
}

func (this *DataStore) BatchRetrive(keys []string, T []any) []interface{} {
	this.commitLock.RLock()
	defer this.commitLock.RUnlock()
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

	if data, err := this.batchFetchPersistentStorage(queryKeys); err == nil { // search for the values that aren't in the cache
		for i, idx := range queryIdxes {
			if data[i] != nil {
				if len(T) > 0 {
					values[idx] = this.decoder(queryKeys[i], data[i], T[i])
				} else {
					values[idx] = this.decoder(queryKeys[i], data[i], nil)
				}
			}
		}
		this.batchAddToCache(this.GetPartitions(keys), keys, values) //update to the local cache and add all the missing values to the cache
	}
	return values
}

func (this *DataStore) Clear() {
	this.partitionIDs = this.partitionIDs[:0]
	this.keyBuffer = this.keyBuffer[:0]
	this.valueBuffer = this.valueBuffer[:0]
	this.encodedBuffer = this.encodedBuffer[:0]
}

func (this *DataStore) Precommit(arg ...interface{}) [32]byte {
	kvs := arg[0].(*CCIndexer).Get().([]interface{})
	keys, values := kvs[0].([]string), kvs[1].([]interface{})

	this.commitLock.Lock()                // Lock the process, only unlock after the final commit is done.
	compressedKeys := common.IfThenDo1st( // Compress the keys if the keyCompressor is available
		this.keyCompressor != nil,
		func() []string { return this.keyCompressor.CompressOnTemp(codec.Strings(keys).Clone()) },
		keys,
	)

	// Encode the keys and values to the buffer so that they can be written to calcualte the root hash.
	encodedBuffer := make([][]byte, len(values))
	for i := 0; i < len(values); i++ {
		if values[i] != nil {
			encodedBuffer[i] = this.encoder(keys[i], values[i])
		}
	}

	this.partitionIDs = append(this.partitionIDs, this.GetPartitions(keys)...)
	this.keyBuffer = append(this.keyBuffer, compressedKeys...)
	this.valueBuffer = append(this.valueBuffer, values...)
	this.encodedBuffer = append(this.encodedBuffer, encodedBuffer...)
	return [32]byte{}
}

// The function calculates the partition id for each key
func (this *DataStore) GetPartitions(keys []string) []uint64 {
	return slice.ParallelTransform(keys, 4, func(i int, k string) uint64 {
		return this.cccache.Hash(k)
	})
}

// Commit the changes to the local cache and the persistent storage
func (this *DataStore) Commit(_ uint64) error {
	for len(this.queue) != 0 {
		fmt.Println("Waiting for the job queue to be empty")
		time.Sleep(100 * time.Millisecond)
	}

	defer this.commitLock.Unlock()                                            // Unlock the process after the final commit is done.
	this.batchAddToCache(this.partitionIDs, this.keyBuffer, this.valueBuffer) // update the local cache

	var err error
	if this.keyCompressor != nil {
		common.ParallelExecute(
			func() { err = this.batchWritePersistentStorage(this.keyBuffer, this.encodedBuffer) }, // Write data back
			func() { this.keyCompressor.Commit() })

	} else {
		err = this.batchWritePersistentStorage(this.keyBuffer, this.encodedBuffer)
	}

	this.cccache.BatchSet(this.keyBuffer, this.valueBuffer) // update the local cache
	this.Clear()
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
