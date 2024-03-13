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
	"strings"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	datastore "github.com/arcology-network/common-lib/storage/datastore"
	platform "github.com/arcology-network/storage-committer/platform"
)

type StoreRouter struct {
	ethDataStore   *EthDataStore
	localDataStore *datastore.DataStore
}

func NewStoreRouter() *StoreRouter {
	return &StoreRouter{
		ethDataStore:   NewParallelEthMemDataStore(),
		localDataStore: datastore.NewDataStore(nil, nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode),
	}
}

func (this *StoreRouter) IfExists(key string) bool {
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
	err1 := this.localDataStore.BatchInject(localKeys, localVals)
	return errors.New(err0.Error() + err1.Error())
}

func (this *StoreRouter) Retrive(key string, v any) (interface{}, error) {
	return this.GetStorage(key).Retrive(key, v)
}

func (this *StoreRouter) RetriveFromStorage(key string, v any) (interface{}, error) {
	return this.GetStorage(key).RetriveFromStorage(key, v)
}

func (this *StoreRouter) BatchRetrive(keys []string, vals []any) []interface{} {
	localKeys, localVals := slice.MoveBothIf(&keys, &vals, func(i int, str string, v any) bool {
		return strings.Contains(str, "/container")
	})

	locals := this.localDataStore.BatchRetrive(localKeys, localVals)
	ethData := this.ethDataStore.BatchRetrive(keys, vals)
	return append(locals, ethData...)
}

func (this *StoreRouter) Precommit(args ...interface{}) [32]byte {
	return this.ethDataStore.Precommit(args)
}

func (this *StoreRouter) Commit(block uint64) error {
	err0 := this.ethDataStore.Commit(block)
	err1 := (this.localDataStore.Commit(block))
	return errors.New(err0.Error() + err1.Error())
}

func (this *StoreRouter) UpdateCacheStats(arg []interface{}) {
	this.ethDataStore.UpdateCacheStats(arg)
	this.localDataStore.UpdateCacheStats(arg)
}

func (this *StoreRouter) Cache(T any) interface{} {
	if common.IsType[*EthDataStore](T) {
		return this.ethDataStore.Cache(T)
	}
	return this.localDataStore.Cache(T)
}

func (this *StoreRouter) Encoder(T any) func(string, interface{}) []byte {
	if common.IsType[*EthDataStore](T) {
		return this.ethDataStore.Encoder(T)
	}
	return this.localDataStore.Encoder(T)
}

func (this *StoreRouter) Decoder(T any) func(string, []byte, any) interface{} {
	if common.IsType[*EthDataStore](T) {
		return this.ethDataStore.Decoder(T)
	}
	return this.localDataStore.Decoder(T)
}

func (this *StoreRouter) Clear() {
	this.ethDataStore.Clear()
	this.localDataStore.Clear()
}

func (this *StoreRouter) Print() {
	this.ethDataStore.Print()
	this.localDataStore.Print()
}

func (this *StoreRouter) CheckSum() [32]byte {
	return [32]byte{}
}

func (this *StoreRouter) Query(string, func(string, string) bool) ([]string, [][]byte, error) {
	return nil, nil, nil
}
