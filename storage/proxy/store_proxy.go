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
	"math"

	ccbadger "github.com/arcology-network/common-lib/storage/badger"
	memdb "github.com/arcology-network/common-lib/storage/memdb"
	"github.com/arcology-network/storage-committer/type/univalue"

	// intf "github.com/arcology-network/storage-committer/interfaces"
	intf "github.com/arcology-network/storage-committer/common"

	stgtypcodec "github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/storage/ethstorage"
	ethstg "github.com/arcology-network/storage-committer/storage/ethstorage"
	livecache "github.com/arcology-network/storage-committer/storage/livecache"
	ccstorage "github.com/arcology-network/storage-committer/storage/livestorage"
	livestg "github.com/arcology-network/storage-committer/storage/livestorage"
)

type StorageProxy struct {
	execCache   *livecache.LiveCache // An object cache for the backend storage, only updated once at the end of the block.
	execStorage *livestg.LiveStorage
	ethStorage  *ethstg.EthDataStore
}

// Cache may also have its storeage, this is the cache only store proxy, no storage.
func NewCacheOnlyStoreProxy() *StorageProxy {
	proxy := &StorageProxy{
		ethStorage: ethstg.NewParallelEthMemDataStore(), //ethstg.NewParallelEthMemDataStore(),
		execStorage: livestg.NewLiveStorage(
			nil,
			stgtypcodec.Codec{}.Encode,
			stgtypcodec.Codec{}.Decode,
		),
	}

	proxy.execCache = livecache.NewLiveCache(math.MaxUint64)
	return proxy
}

func NewMemDBStoreProxy() *StorageProxy {
	proxy := NewCacheOnlyStoreProxy()
	proxy.execStorage.SetDB(memdb.NewMemoryDB())
	return proxy
}

func NewLevelDBStoreProxy(dbpath string) *StorageProxy {
	proxy := &StorageProxy{
		ethStorage: ethstg.NewLevelDBDataStore(dbpath), //ethstg.NewParallelEthMemDataStore(),
		execStorage: livestg.NewLiveStorage(
			// memdb.NewMemoryDB(),
			ccbadger.NewBadgerDB(dbpath+"_badager"),
			// ccbadger.NewParaBadgerDB(dbpath+"_pbadager", common.Remainder),
			stgtypcodec.Codec{}.Encode,
			stgtypcodec.Codec{}.Decode,
		),
	}
	proxy.execCache = livecache.NewLiveCache(math.MaxUint64)
	return proxy
}

// NewStoreProxyPersistentDB creates a new storage proxy with a persistent databases
// func NewTestLevelDBStoreProxy() *StorageProxy {
// 	return NewLevelDBStoreProxy("/tmp")
// }

func (this *StorageProxy) EnableCache() *StorageProxy {
	this.execCache.Enable()
	return this
}

func (this *StorageProxy) DisableCache() *StorageProxy {
	this.execCache.Disable()
	return this
}

func (this *StorageProxy) ClearExecCache() { this.execCache.Clear() }

func (this *StorageProxy) ExecCache() *livecache.LiveCache { return this.execCache }
func (this *StorageProxy) ExecStore() *livestg.LiveStorage { return this.execStorage } // Arcology storage

// Check if the key exists in th storage.
func (this *StorageProxy) RetriveFromStorage(key string, T any) (any, error) {
	if v, ok := this.execCache.Get(key); ok { // Check the cache first
		return v, nil
	}
	return this.execStorage.Retrive(key, T)
}

func (this *StorageProxy) EthStore() *ethstg.EthDataStore { return this.ethStorage } // Eth storage

func (this *StorageProxy) Preload(data []byte) any {
	return this.ethStorage.Preload(data)
}

// Check if the key exists in the source, which can be a cache or a storage.
func (this *StorageProxy) IfExists(key string) bool {
	if _, ok := this.execCache.Get(key); ok { // Check the cache first
		return true
	}
	return this.execStorage.IfExists(key)
}

// Directly inject the value into the storage, on for the concurrent container storage
func (this *StorageProxy) Inject(key string, v any) error {
	return this.execStorage.Inject(key, v)
}

func (this *StorageProxy) Retrive(key string, v any) (any, error) {
	if retv, ok := this.execCache.Get(key); ok { // Get from cache first
		return retv, nil
	}
	return this.execStorage.Retrive(key, v)
}

// Get the stores that can be
func (this *StorageProxy) GetWriters() []intf.AsyncWriter[*univalue.Univalue] {
	return []intf.AsyncWriter[*univalue.Univalue]{
		livecache.NewLiveCacheWriter(this.execCache, -1),
		ethstorage.NewEthStorageWriter(this.ethStorage, -1),
		ccstorage.NewLiveStorageWriter(this.execStorage, -1),
	}
}
