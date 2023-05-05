package ccdb

import (
	"crypto/sha256"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
)

type TransientDB struct {
	*cachedstorage.DataStore
	parent ccurlcommon.DatastoreInterface
}

func NewTransientDB(parent ccurlcommon.DatastoreInterface) ccurlcommon.DatastoreInterface {
	return &TransientDB{
		DataStore: cachedstorage.NewDataStore(),
		parent:    parent,
	}
}

func (db *TransientDB) Query(pattern string, condition func(string, string) bool) ([]string, [][]byte, error) {
	return []string{}, [][]byte{}, nil
}

func (db *TransientDB) Inject(path string, v interface{}) {
	db.DataStore.Inject(path, v)
}

func (this *TransientDB) Checksum() [32]byte {
	return this.DataStore.Checksum()
}

func (db *TransientDB) Retrive(path string) (interface{}, error) {
	v, err := db.DataStore.Retrive(path)
	if err != nil {
		//return nil, err
	}
	if v == nil {
		v, err = db.parent.Retrive(path)
		if err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (db *TransientDB) BatchRetrive(paths []string) []interface{} {
	queryKeys := make([]string, 0, len(paths))
	queryIdxes := make([]int, 0, len(paths))
	values := db.DataStore.BatchRetrive(paths)
	for i := 0; i < len(paths); i++ {
		if values[i] == nil {
			queryKeys = append(queryKeys, paths[i])
			queryIdxes = append(queryIdxes, i)
		}
	}

	if len(queryKeys) == 0 { // No missing values
		return values
	}
	queryvalues := db.parent.BatchRetrive(queryKeys)
	for i, idx := range queryIdxes {
		values[idx] = queryvalues[i]
	}

	return values
}

func (db *TransientDB) Precommit(paths []string, dict interface{}) {
	db.DataStore.Precommit(paths, dict)
}

func (db *TransientDB) Commit() error {
	return db.DataStore.Commit()
}

func (this *TransientDB) Print() {
	this.DataStore.Print()
}
func (this *TransientDB) CheckSum() [32]byte {
	psum := this.parent.CheckSum()
	tsum := this.DataStore.CheckSum()
	datas := []byte{}
	datas = append(datas, psum[:]...)
	datas = append(datas, tsum[:]...)
	return sha256.Sum256(datas)
}

func (this *TransientDB) Dump() ([]string, []interface{}) {
	pkeys, pvals := this.parent.Dump()
	keys, vals := this.DataStore.Dump()

	return append(pkeys, keys...), append(pvals, vals...)
}

func (this *TransientDB) UpdateCacheStats(vals []interface{}) {
	this.DataStore.UpdateCacheStats(vals)
}

func (this *TransientDB) CacheRetrive(key string, valueTransformer func(interface{}) interface{}) (interface{}, error) {
	if v, err := this.DataStore.CacheRetrive(key, valueTransformer); v != nil {
		return v, err
	}
	return this.parent.CacheRetrive(key, valueTransformer)
}
