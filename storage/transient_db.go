package storage

import (
	"crypto/sha256"
	"math"

	datastore "github.com/arcology-network/common-lib/storage/datastore"
	memdb "github.com/arcology-network/common-lib/storage/memdb"
	policy "github.com/arcology-network/common-lib/storage/policy"
	"github.com/arcology-network/storage-committer/interfaces"
)

type TransientDB struct {
	*datastore.DataStore
	readonlyParent interfaces.Datastore
}

func NewTransientDB(readonlyParent interfaces.Datastore) interfaces.Datastore {
	return &TransientDB{
		DataStore: datastore.NewDataStore(
			nil,
			policy.NewCachePolicy(math.MaxUint64, 1),
			memdb.NewMemoryDB(),
			Rlp{}.Encode,
			Rlp{}.Decode,
		),
		readonlyParent: readonlyParent,
	}
}

// placeholder, TransientDB does not need this at all.
func (this *TransientDB) Preload([]byte) interface{} { return nil }

func (this *TransientDB) WriteEthTries(...interface{}) [32]byte { return [32]byte{} }

// placeholder, TransientDB does not need this at all.
func (this *TransientDB) Cache(any) interface{} { return nil }

func (this *TransientDB) Query(pattern string, condition func(string, string) bool) ([]string, [][]byte, error) {
	return []string{}, [][]byte{}, nil
}

func (this *TransientDB) Inject(path string, v interface{}) error {
	return this.DataStore.Inject(path, v)
}

func (this *TransientDB) Precommit(arg ...interface{}) [32]byte {
	return this.DataStore.Precommit(arg...)
}

func (this *TransientDB) IfExists(key string) bool {
	return this.DataStore.IfExists(key) || this.readonlyParent.IfExists(key)
}

func (this *TransientDB) Commit(_ uint64) error { return this.DataStore.Commit(0) }

// func (this *TransientDB) Checksum() [32]byte { return this.DataStore.Checksum() }
func (this *TransientDB) Print() { this.DataStore.Print() }
func (this *TransientDB) Buffers() ([]string, []interface{}, [][]byte) {
	return this.DataStore.Buffers()
}

func (this *TransientDB) Retrive(path string, T any) (interface{}, error) {
	v, err := this.DataStore.Retrive(path, T)
	if err != nil {
		return nil, err
	}

	if v == nil {
		v, err = this.readonlyParent.Retrive(path, T)
		if err != nil {
			return nil, err
		}
	}
	return v, nil
}

// Placeholder, skip the cache and read from storage directly
func (this *TransientDB) RetriveFromStorage(path string, T any) (interface{}, error) {
	return this.Retrive(path, T)
}

func (this *TransientDB) BatchRetrive(paths []string, T []any) []interface{} {
	queryKeys := make([]string, 0, len(paths))
	queryIdxes := make([]int, 0, len(paths))
	values := this.DataStore.BatchRetrive(paths, T)
	for i := 0; i < len(paths); i++ {
		if values[i] == nil {
			queryKeys = append(queryKeys, paths[i])
			queryIdxes = append(queryIdxes, i)
		}
	}

	if len(queryKeys) == 0 { // No missing values
		return values
	}
	queryvalues := this.readonlyParent.BatchRetrive(queryKeys, T)
	for i, idx := range queryIdxes {
		values[idx] = queryvalues[i]
	}

	return values
}

func (this *TransientDB) CheckSum() [32]byte {
	psum := this.readonlyParent.CheckSum()
	tsum := this.DataStore.CheckSum()
	datas := []byte{}
	datas = append(datas, psum[:]...)
	datas = append(datas, tsum[:]...)
	return sha256.Sum256(datas)
}

// func (this *TransientDB) Dump() ([]string, []interface{}) {
// 	pkeys, pvals := this.readonlyParent.(*datastore.DataStore).KVs()
// 	keys, vals := this.DataStore.KVs()
// 	return append(pkeys, keys...), append(pvals, vals...)
// }

func (this *TransientDB) UpdateCacheStats(vals []interface{}) {
	this.DataStore.UpdateCacheStats(vals)
}
