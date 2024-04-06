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
	"github.com/arcology-network/common-lib/exp/associative"
	memdb "github.com/arcology-network/common-lib/storage/memdb"
	policy "github.com/arcology-network/common-lib/storage/policy"
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
	ccstg "github.com/arcology-network/storage-committer/storage/ccstorage"
	ethstg "github.com/arcology-network/storage-committer/storage/ethstorage"
	"github.com/arcology-network/storage-committer/univalue"
)

type StorageProxy struct {
	objectCache  *ReadCache // An object cache for the backend storage, only updated once at the end of the block.
	ethDataStore *ethstg.EthDataStore
	ccDataStore  *ccstg.DataStore
}

func NewStoreProxy() *StorageProxy {
	proxy := &StorageProxy{
		ethDataStore: ethstg.NewParallelEthMemDataStore(),
		ccDataStore: ccstg.NewDataStore(
			nil,
			policy.NewCachePolicy(0, 1), // Don't cache anything in the underlying storage, the cache is managed by the router
			memdb.NewMemoryDB(),
			platform.Codec{}.Encode, platform.Codec{}.Decode),
	}
	proxy.objectCache = NewReadCache(proxy)

	// Start the listeners
	// go proxy.objectCache.Start()
	// go proxy.ethDataStore.Start()
	// go proxy.ccDataStore.Start()
	return proxy
}

func (this *StorageProxy) Cache(any) interface{} {
	return this.objectCache
}

func (this *StorageProxy) EnableCache() *StorageProxy {
	this.objectCache.Enable()
	return this
}

func (this *StorageProxy) DisableCache() *StorageProxy {
	this.objectCache.Disable()
	return this
}

func (this *StorageProxy) ClearCache() { this.objectCache.Clear() }

func (this *StorageProxy) EthStore() *ethstg.EthDataStore { return this.ethDataStore } // Eth storage
func (this *StorageProxy) CCStore() *ccstg.DataStore      { return this.ccDataStore }  // Arcology storage

func (this *StorageProxy) Preload(data []byte) interface{} {
	return this.ethDataStore.Preload(data)
}

func (this *StorageProxy) IfExists(key string) bool {
	if _, ok := this.objectCache.Get(key); ok { // Check the cache first
		return true
	}
	return this.GetStorage(key).IfExists(key)
}

func (this *StorageProxy) Inject(key string, v any) error {
	return this.GetStorage(key).Inject(key, v)
}

func (this *StorageProxy) Retrive(key string, v any) (interface{}, error) {
	if v, ok := this.objectCache.Get(key); ok { // Get from cache first
		return *v, nil
	}
	return this.GetStorage(key).Retrive(key, v)
}

// Placeholders for the storage interface
func (this *StorageProxy) AsyncPrecommit(...interface{})          {}
func (this *StorageProxy) Precommit(args ...interface{}) [32]byte { return this.ethDataStore.Root() }
func (this *StorageProxy) Commit(blockNum uint64) error           { return nil }

// Get the stores that can be
func (this *StorageProxy) Committable() []*associative.Pair[intf.Indexer[*univalue.Univalue], []intf.CommittableStore] {
	bufferPair := &associative.Pair[intf.Indexer[*univalue.Univalue], []intf.CommittableStore]{
		First: NewIndexer(this),
		Second: []intf.CommittableStore{
			this.objectCache,
		}}

	ethPair := &associative.Pair[intf.Indexer[*univalue.Univalue], []intf.CommittableStore]{
		First:  ethstg.NewIndexer(this),
		Second: []intf.CommittableStore{this.EthStore()},
	}

	ccPair := &associative.Pair[intf.Indexer[*univalue.Univalue], []intf.CommittableStore]{
		First:  ccstg.NewIndexer(this),
		Second: []intf.CommittableStore{this.CCStore()},
	}

	return []*associative.Pair[intf.Indexer[*univalue.Univalue], []intf.CommittableStore]{
		bufferPair,
		ethPair,
		ccPair,
	}
}
