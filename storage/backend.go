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
	datastore "github.com/arcology-network/common-lib/storage/datastore"
	memdb "github.com/arcology-network/common-lib/storage/memdb"
	policy "github.com/arcology-network/common-lib/storage/policy"
	ethstg "github.com/arcology-network/storage-committer/ethstorage"
	intf "github.com/arcology-network/storage-committer/interfaces"
	platform "github.com/arcology-network/storage-committer/platform"
)

type StoreBackend[T any] struct {
	backendDbs []intf.Datastore
	// ethDataStore *ethstg.EthDataStore
	// ccDataStore  *datastore.DataStore
}

func NewLocalStoreBackend[T any]() *StoreBackend[T] {
	ethDataStore := ethstg.NewParallelEthMemDataStore()
	ccDataStore := datastore.NewDataStore(
		nil,
		policy.NewCachePolicy(0, 1), // Don't cache anything in the underlying storage, the cache is managed by the router
		memdb.NewMemoryDB(),
		platform.Codec{}.Encode, platform.Codec{}.Decode)

	return &StoreBackend[T]{
		backendDbs: []intf.Datastore{ethDataStore, ccDataStore},
	}
}

// New creates a new StateCommitter instance.
func (this *StoreBackend[T]) Index(vals ...T) {
	// for i, v := range this.backendDbs {
	// 	// this.backendDbs[i].Index(v)
	// }

}
