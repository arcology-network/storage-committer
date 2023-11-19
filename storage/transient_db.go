package ccdb

import (
	"crypto/sha256"
	"math"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type TransientDB struct {
	*cachedstorage.DataStore
	readonlyParent interfaces.Datastore
}

func NewTransientDB(readonlyParent interfaces.Datastore) interfaces.Datastore {
	return &TransientDB{
		DataStore: cachedstorage.NewDataStore(
			nil,
			cachedstorage.NewCachePolicy(math.MaxUint64, 1), cachedstorage.NewMemDB(), Rlp{}.Encode, Rlp{}.Decode,
		),
		readonlyParent: readonlyParent,
	}
}

func (this *TransientDB) Query(pattern string, condition func(string, string) bool) ([]string, [][]byte, error) {
	return []string{}, [][]byte{}, nil
}

func (this *TransientDB) Inject(path string, v interface{}) error {
	return this.DataStore.Inject(path, v)
}

func (this *TransientDB) Precommit(paths []string, dict interface{}) [32]byte {
	return this.DataStore.Precommit(paths, dict)
}

func (this *TransientDB) IfExists(key string) bool {
	return this.DataStore.IfExists(key) || this.readonlyParent.IfExists(key)
}

func (this *TransientDB) Commit() error { return this.DataStore.Commit() }

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

func (this *TransientDB) Dump() ([]string, []interface{}) {
	pkeys, pvals := this.readonlyParent.Dump()
	keys, vals := this.DataStore.Dump()

	return append(pkeys, keys...), append(pvals, vals...)
}

func (this *TransientDB) UpdateCacheStats(vals []interface{}) {
	this.DataStore.UpdateCacheStats(vals)
}
