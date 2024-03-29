/*
 *   Copyright (c) 2024 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package proxy

import (
	"errors"

	cache "github.com/arcology-network/common-lib/cache"
	"github.com/arcology-network/common-lib/common"
	memdb "github.com/arcology-network/common-lib/storage/memdb"
	policy "github.com/arcology-network/common-lib/storage/policy"
	"github.com/arcology-network/storage-committer/commutative"
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
	datastore "github.com/arcology-network/storage-committer/storage/ccstorage"
	ethstg "github.com/arcology-network/storage-committer/storage/ethstorage"
)

type StoreProxy struct {
	objectCache  *cache.ReadCache[string, intf.Type] // Cache shared by all storage
	ethDataStore *ethstg.EthDataStore
	ccDataStore  *datastore.DataStore[string, intf.Type]
}

func NewStoreProxy() *StoreProxy {
	return &StoreProxy{
		objectCache: cache.NewReadCache[string, intf.Type](
			4096, // 4096 shards to avoid lock contention
			func(v intf.Type) bool { return v == nil },
		),
		ethDataStore: ethstg.NewParallelEthMemDataStore(),
		ccDataStore: datastore.NewDataStore[string, intf.Type](
			nil,
			policy.NewCachePolicy(0, 1), // Don't cache anything in the underlying storage, the cache is managed by the router
			memdb.NewMemoryDB(),
			platform.Codec{}.Encode, platform.Codec{}.Decode),
	}
}

func (this *StoreProxy) Cache(any) interface{}     { return this.objectCache }
func (this *StoreProxy) EnableCache() *StoreProxy  { this.objectCache.Enable(); return this }
func (this *StoreProxy) DisableCache() *StoreProxy { this.objectCache.Disable(); return this }
func (this *StoreProxy) ClearCache()               { this.objectCache.Clear() }

func (this *StoreProxy) EthStore() *ethstg.EthDataStore                   { return this.ethDataStore } // Eth storage
func (this *StoreProxy) CCStore() *datastore.DataStore[string, intf.Type] { return this.ccDataStore }  // Arcology storage

func (this *StoreProxy) Precommit(args ...interface{}) [32]byte { return [32]byte{} }

func (this *StoreProxy) Preload(data []byte) interface{} {
	return this.ethDataStore.Preload(data)
}

func (this *StoreProxy) IfExists(key string) bool {
	if _, ok := this.objectCache.Get(key); ok { // Check the cache first
		return true
	}
	return this.GetStorage(key).IfExists(key)
}

func (this *StoreProxy) Inject(key string, v any) error {
	return this.GetStorage(key).Inject(key, v)
}

func (this *StoreProxy) Retrive(key string, v any) (interface{}, error) {
	if v, ok := this.objectCache.Get(key); ok { // Get from cache first
		return *v, nil
	}
	return this.GetStorage(key).Retrive(key, v)
}

func (this *StoreProxy) Commit(blockNum uint64) error {
	err0 := this.ethDataStore.Commit(blockNum)
	err1 := (this.ccDataStore.Commit(blockNum))
	return errors.New(err0.Error() + err1.Error())
}

// Update the object cache.
func (this *StoreProxy) RefreshCache(blockNum uint64, dirtyKeys []string, dirtyVals []intf.Type) {
	for _, v := range dirtyVals {
		if common.IsType[*commutative.Uint64](v) && v.(*commutative.Uint64).Delta().(uint64) != 0 {
			panic("Error: Delta value should not be in the dirtyVals")
		}
	}
	this.objectCache.Commit(dirtyKeys, dirtyVals)
}
