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
	"runtime"
	"strings"

	cache "github.com/arcology-network/common-lib/cache"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	datastore "github.com/arcology-network/common-lib/storage/datastore"
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
	"github.com/cespare/xxhash/v2"
)

type StoreRouter struct {
	cache       *cache.ReadCache[string, intf.Type] // Cache shared by all storage
	cacheActive bool

	ethDataStore *EthDataStore
	ccDataStore  *datastore.DataStore
}

func NewHybirdStore() *StoreRouter {
	return &StoreRouter{
		cache: cache.NewReadCache[string, intf.Type](
			4096, // 4096 shards to avoid lock contention
			func(k string) uint64 { return xxhash.Sum64String(k) },
			func(v intf.Type) bool { return v == nil },
		),
		cacheActive:  true,
		ethDataStore: NewParallelEthMemDataStore(),
		ccDataStore: datastore.NewDataStore(
			nil,
			nil,
			nil,
			platform.Codec{}.Encode, platform.Codec{}.Decode),
	}
}

func (this *StoreRouter) Cache(any) interface{}     { return this.cache }
func (this *StoreRouter) ActivateCache(active bool) { this.cacheActive = active }
func (this *StoreRouter) ClearCache()               { this.cache.Clear() }

func (this *StoreRouter) EthStore() *EthDataStore       { return this.ethDataStore } // Eth storage
func (this *StoreRouter) CCStore() *datastore.DataStore { return this.ccDataStore }  // Arcology storage

func (this *StoreRouter) Precommit(args ...interface{}) [32]byte { return [32]byte{} }

func (this *StoreRouter) Preload(data []byte) interface{} {
	return this.ethDataStore.Preload(data)
	// this.ccDataStore.Preload(data)
	// return nil
}

func (this *StoreRouter) IfExists(key string) bool {
	if _, ok := this.cache.Get(key); ok { // Check the cache first
		return true
	}
	return this.GetStorage(key).IfExists(key)
}

func (this *StoreRouter) Inject(key string, v any) error {
	return this.GetStorage(key).Inject(key, v)
}

func (this *StoreRouter) BatchInject(key []string, vals []any) error {
	localKeys, localVals := slice.MoveBothIf(&key, &vals, func(i int, str string, v any) bool {
		return strings.Contains(str, "/container")
	})

	err0 := this.ethDataStore.BatchInject(key, vals)
	err1 := this.ccDataStore.BatchInject(localKeys, localVals)
	return errors.New(err0.Error() + err1.Error())
}

func (this *StoreRouter) Retrive(key string, v any) (interface{}, error) {
	if v, ok := this.cache.Get(key); ok { // Get from cache first
		return *v, nil
	}
	return this.GetStorage(key).Retrive(key, v)
}

func (this *StoreRouter) RetriveFromStorage(key string, v any) (interface{}, error) {
	return this.GetStorage(key).RetriveFromStorage(key, v)
}

func (this *StoreRouter) BatchRetrive(keys []string, vals []any) []interface{} {
	return slice.ParallelTransform(keys, runtime.NumCPU(), func(i int, k string) interface{} {
		v, _ := this.cache.Get(k)
		return v
	})
}

func (this *StoreRouter) Commit(blockNum uint64) error {
	err0 := this.ethDataStore.Commit(blockNum)
	err1 := (this.ccDataStore.Commit(blockNum))
	return errors.New(err0.Error() + err1.Error())
}

// Update the object cache.
func (this *StoreRouter) RefreshCache(blockNum uint64, dirtyKeys []string, dirtyVals []intf.Type) {
	if this.cacheActive {
		this.cache.Commit(dirtyKeys, dirtyVals)
	}
}

func (this *StoreRouter) UpdateCacheStats(arg []interface{}) {
	this.ethDataStore.UpdateCacheStats(arg)
	this.ccDataStore.UpdateCacheStats(arg)
}

func (this *StoreRouter) Encoder(T any) func(string, interface{}) []byte {
	if common.IsType[*EthDataStore](T) {
		return this.ethDataStore.Encoder(T)
	}
	return this.ccDataStore.Encoder(T)
}

func (this *StoreRouter) Decoder(T any) func(string, []byte, any) interface{} {
	if common.IsType[*EthDataStore](T) {
		return this.ethDataStore.Decoder(T)
	}
	return this.ccDataStore.Decoder(T)
}

func (this *StoreRouter) Clear() {
	this.ethDataStore.Clear()
	this.ccDataStore.Clear()
}

func (this *StoreRouter) Print() {
	this.ethDataStore.Print()
	this.ccDataStore.Print()
}

func (this *StoreRouter) CheckSum() [32]byte {
	return [32]byte{}
}

func (this *StoreRouter) Query(string, func(string, string) bool) ([]string, [][]byte, error) {
	return nil, nil, nil
}
