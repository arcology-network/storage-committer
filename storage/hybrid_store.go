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

package storage

import (
	"errors"

	cache "github.com/arcology-network/common-lib/cache"
	"github.com/arcology-network/common-lib/common"
	datastore "github.com/arcology-network/common-lib/storage/datastore"
	memdb "github.com/arcology-network/common-lib/storage/memdb"
	policy "github.com/arcology-network/common-lib/storage/policy"
	"github.com/arcology-network/storage-committer/commutative"
	ethstg "github.com/arcology-network/storage-committer/ethstorage"
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
)

type StoreRouter struct {
	objectCache  *cache.ReadCache[string, intf.Type] // Cache shared by all storage
	ethDataStore *ethstg.EthDataStore
	ccDataStore  *datastore.DataStore
}

func NewHybirdStore() *StoreRouter {
	return &StoreRouter{
		objectCache: cache.NewReadCache[string, intf.Type](
			4096, // 4096 shards to avoid lock contention
			func(v intf.Type) bool { return v == nil },
		),
		ethDataStore: ethstg.NewParallelEthMemDataStore(),
		ccDataStore: datastore.NewDataStore(
			nil,
			policy.NewCachePolicy(0, 1), // Don't cache anything in the underlying storage, the cache is managed by the router
			memdb.NewMemoryDB(),
			platform.Codec{}.Encode, platform.Codec{}.Decode),
	}
}

func (this *StoreRouter) Cache(any) interface{}      { return this.objectCache }
func (this *StoreRouter) EnableCache() *StoreRouter  { this.objectCache.Enable(); return this }
func (this *StoreRouter) DisableCache() *StoreRouter { this.objectCache.Disable(); return this }
func (this *StoreRouter) ClearCache()                { this.objectCache.Clear() }

func (this *StoreRouter) EthStore() *ethstg.EthDataStore { return this.ethDataStore } // Eth storage
func (this *StoreRouter) CCStore() *datastore.DataStore  { return this.ccDataStore }  // Arcology storage

func (this *StoreRouter) Precommit(args ...interface{}) [32]byte { return [32]byte{} }

func (this *StoreRouter) Preload(data []byte) interface{} {
	return this.ethDataStore.Preload(data)
}

func (this *StoreRouter) IfExists(key string) bool {
	if _, ok := this.objectCache.Get(key); ok { // Check the cache first
		return true
	}
	return this.GetStorage(key).IfExists(key)
}

func (this *StoreRouter) Inject(key string, v any) error {
	return this.GetStorage(key).Inject(key, v)
}

func (this *StoreRouter) Retrive(key string, v any) (interface{}, error) {
	if v, ok := this.objectCache.Get(key); ok { // Get from cache first
		return *v, nil
	}
	return this.GetStorage(key).Retrive(key, v)
}

func (this *StoreRouter) Commit(blockNum uint64) error {
	err0 := this.ethDataStore.Commit(blockNum)
	err1 := (this.ccDataStore.Commit(blockNum))
	return errors.New(err0.Error() + err1.Error())
}

// Update the object cache.
func (this *StoreRouter) RefreshCache(blockNum uint64, dirtyKeys []string, dirtyVals []intf.Type) {
	for _, v := range dirtyVals {
		if common.IsType[*commutative.Uint64](v) && v.(*commutative.Uint64).Delta().(uint64) != 0 {
			panic("Error: Delta value should not be in the dirtyVals")
		}
	}
	this.objectCache.Commit(dirtyKeys, dirtyVals)
}
