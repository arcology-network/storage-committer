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
package committertest

import (
	"testing"
)

func TestWriteCachePool(t *testing.T) {
	// store := stgproxy.NewStoreProxy()
	// caches := make([]*cache.WriteCache, 10000)

	// t0 := time.Now()
	// for i := 0; i < len(caches); i++ {
	// 	caches[i] = cache.NewSequentialWriteCache(store)
	// }
	// t.Logf("Time to create 10000 native map caches: %v", time.Since(t0))

	// t0 = time.Now()
	// for i := 0; i < len(caches); i++ {
	// 	ccmap := mapi.NewConcurrentMap(int(16),
	// 		func(v *univalue.Univalue) bool {
	// 			return v == nil
	// 		},
	// 		func(k string) uint64 {
	// 			return xxhash.Sum64String(k)
	// 		})

	// 	pool := mempool.NewMempool(1, 16, func() *univalue.Univalue { return new(univalue.Univalue) }, (&univalue.Univalue{}).Reset)
	// 	caches[i] = cache.NewWriteCache(store, ccmap, pool)
	// }
	// t.Logf("Time to create 10000 ccmap caches: %v", time.Since(t0))

	// datastore := ccurlstorage.NewParallelEthMemDataStore()

	// // create a pool of 16 write caches.
	// writeCachePool := mempool.NewMempool[*WriteCache](16, 1, func() *WriteCache {
	// 	return NewSequentialWriteCache(datastore, 32, 1)
	// })

}
