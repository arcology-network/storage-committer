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
	memdb "github.com/arcology-network/common-lib/storage/memdb"
	policy "github.com/arcology-network/common-lib/storage/policy"
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/storage/ccstorage"
	ccstg "github.com/arcology-network/storage-committer/storage/ccstorage"
	"github.com/arcology-network/storage-committer/storage/ethstorage"
	ethstg "github.com/arcology-network/storage-committer/storage/ethstorage"
	"github.com/arcology-network/storage-committer/univalue"
)

type StorageProxy struct {
	unifiedCache *ReadCache // An object cache for the backend storage, only updated once at the end of the block.
	ethDataStore *ethstg.EthDataStore
	ccDataStore  *ccstg.DataStore
}

func NewMemDBStoreProxy() *StorageProxy {
	proxy := &StorageProxy{
		ethDataStore: ethstg.NewParallelEthMemDataStore(), //ethstg.NewParallelEthMemDataStore(),
		ccDataStore: ccstg.NewDataStore(
			policy.NewCachePolicy(0, 1), // Don't cache anything in the underlying storage, the cache is managed by the router
			memdb.NewMemoryDB(),
			platform.Codec{}.Encode,
			platform.Codec{}.Decode,
		),
	}
	proxy.unifiedCache = NewReadCache(proxy)
	return proxy
}

func NewLevelDBStoreProxy(dbpath string) *StorageProxy {
	proxy := &StorageProxy{
		ethDataStore: ethstg.NewLevelDBDataStore(dbpath), //ethstg.NewParallelEthMemDataStore(),
		ccDataStore: ccstg.NewDataStore(
			policy.NewCachePolicy(0, 1), // Don't cache anything in the underlying storage, the cache is managed by the router
			memdb.NewMemoryDB(),
			platform.Codec{}.Encode,
			platform.Codec{}.Decode,
		),
	}
	proxy.unifiedCache = NewReadCache(proxy)
	return proxy
}

// NewStoreProxyPersistentDB creates a new storage proxy with a persistent databases
func NewTestLevelDBStoreProxy() *StorageProxy {
	return NewLevelDBStoreProxy("/tmp")
}

func (this *StorageProxy) Cache() interface{} {
	return this.unifiedCache
}

func (this *StorageProxy) EnableCache() *StorageProxy {
	this.unifiedCache.Enable()
	return this
}

func (this *StorageProxy) DisableCache() *StorageProxy {
	this.unifiedCache.Disable()
	return this
}

func (this *StorageProxy) ClearCache() { this.unifiedCache.Clear() }

func (this *StorageProxy) EthStore() *ethstg.EthDataStore { return this.ethDataStore } // Eth storage
func (this *StorageProxy) CCStore() *ccstg.DataStore      { return this.ccDataStore }  // Arcology storage

func (this *StorageProxy) Preload(data []byte) interface{} {
	return this.ethDataStore.Preload(data)
}

func (this *StorageProxy) IfExists(key string) bool {
	if _, ok := this.unifiedCache.Get(key); ok { // Check the cache first
		return true
	}
	return this.ccDataStore.IfExists(key)
}

// Directly inject the value into the storage, on for the concurrent container storage
func (this *StorageProxy) Inject(key string, v any) error {
	return this.ccDataStore.Inject(key, v)
}

func (this *StorageProxy) Retrive(key string, v any) (interface{}, error) {
	if v, ok := this.unifiedCache.Get(key); ok { // Get from cache first
		return *v, nil
	}
	return this.ccDataStore.Retrive(key, v)
}

// Placeholders for the storage interface
func (this *StorageProxy) Precommit([]uint32) [32]byte  { return this.ethDataStore.Root() }
func (this *StorageProxy) Commit(blockNum uint64) error { return nil }

// Get the stores that can be
func (this *StorageProxy) GetWriters() []intf.AsyncWriter[*univalue.Univalue] {
	return []intf.AsyncWriter[*univalue.Univalue]{
		NewAsyncWriter(this.unifiedCache, -1),
		ethstorage.NewAsyncWriter(this.ethDataStore, -1),
		ccstorage.NewAsyncWriter(this.ccDataStore, -1),
	}
}
