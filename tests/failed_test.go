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
	"fmt"
	"testing"
	"time"

	cache "github.com/arcology-network/common-lib/cache"
	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/slice"
	eucache "github.com/arcology-network/eu/cache"
	stgcommitter "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/commutative"
	importer "github.com/arcology-network/storage-committer/importer"
	"github.com/arcology-network/storage-committer/interfaces"
	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
	platform "github.com/arcology-network/storage-committer/platform"
	univalue "github.com/arcology-network/storage-committer/univalue"
	"github.com/holiman/uint256"
)

func TestPathReadAndWriteBatchCache(b *testing.T) {
	store := chooseDataStore()
	writeCache := NewWriteCacheWithAcounts(store, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	bob := BobAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	keys := RandomKeys(0, 2)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
			b.Error(err)
		}
	}

	committer := stgcommitter.NewStorageCommitter(store).Import(univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})).Sort()
	committer.Precommit([]uint32{0})
	committer.Commit()

	for i := 0; i < len(keys); i++ {
		v, ok := store.Cache(nil).(*cache.ReadCache[string, interfaces.Type]).Get("blcc://eth1.0/account/" + alice + "/storage/container/ctrn-0/alice-elem-" + keys[i])
		if typedv, _, _ := (*(v)).Get(); !ok || typedv != int64(i) {
			b.Error("not found")
		}
	}

	// Rewrite the same keys
	writeCache = NewWriteCacheWithAcounts(store)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i+9999))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)+9999))); err != nil {
			b.Error(err)
		}
	}

	trans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})
	committer = stgcommitter.NewStorageCommitter(store).Import(trans).Sort()
	committer.Precommit([]uint32{0})
	committer.Commit()

	for i := 0; i < len(keys); i++ {
		v, ok := store.Cache(nil).(*cache.ReadCache[string, interfaces.Type]).Get("blcc://eth1.0/account/" + alice + "/storage/container/ctrn-0/alice-elem-" + keys[i])
		if typedv, _, _ := (*(v)).Get(); !ok || typedv != int64(i+9999) {
			b.Error("not found")
		}
	}
}

func TestPathReadAndWriteBatch(b *testing.T) {
	store := chooseDataStore()
	writeCache := NewWriteCacheWithAcounts(store, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	bob := BobAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	keys := RandomKeys(0, 100)
	t0 := time.Now()
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("First Write time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < len(keys); i++ {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}

		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i+len(keys)) {
			b.Error(err)
		}
	}
	fmt.Println("First Read time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("1. New Path Write time:", time.Since(t0))

	t0 = time.Now()
	committer := stgcommitter.NewStorageCommitter(store).Import(univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})).Sort()
	committer.Precommit([]uint32{0})
	committer.Commit()
	fmt.Println("Commit time:", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < len(keys); i++ {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}

		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i+len(keys)) {
			b.Error(err)
		}
	}
	fmt.Println("2. First read time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	keys = RandomKeys(len(keys), len(keys)+10)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("Second Write time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	keys = RandomKeys(100000, 100000*2)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("2. New Path Write time:", len(keys), "keys in", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < len(keys); i++ {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/alice-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}
	}
	fmt.Println("2. Second read time:", len(keys), "keys in", time.Since(t0))

	t0 = time.Now()
	keys = RandomKeys(100000*2, 100000*2+20)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println(" New Path Write time:", len(keys), "keys in", time.Since(t0))
}

func TestPathReadAndWrites(b *testing.T) {
	store := chooseDataStore()
	writeCache := NewWriteCacheWithAcounts(store, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	bob := BobAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	keys := RandomKeys(0, 100)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
			b.Error(err)
		}
	}

	committer := stgcommitter.NewStorageCommitter(store).Import(univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})).Sort()
	committer.Precommit([]uint32{0})
	committer.Commit()

	for i := 0; i < len(keys); i++ {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}

		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i+len(keys)) {
			b.Error(err)
		}
	}

	keys = RandomKeys(101, 200)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
			b.Error(err)
		}
	}

	for i := 0; i < len(keys); i++ {
		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i) {
			b.Error(err)
		}

		if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], new(noncommutative.Int64)); v == nil ||
			v.(int64) != int64(i+len(keys)) {
			b.Error(err)
		}
	}
}

func TestPathReadAndWritesPath(b *testing.T) {
	store := chooseDataStore()

	writeCache := eucache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	bob := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, bob); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	keys1, key2 := RandomKey(1), RandomKey(2)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+keys1, commutative.NewUnboundedU256()); err != nil {
		b.Error(err)
	}

	u256 := new(commutative.U256).NewBoundedU256FromUint64(111, 0, 0, 999, true)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+key2, u256); err != nil {
		b.Error(err)
	}

	committer := stgcommitter.NewStorageCommitter(store).Import(univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})).Sort()
	committer.Precommit([]uint32{0})
	committer.Commit()

	writeCache = eucache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if typedv, _, _ := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", new(commutative.Path)); typedv == nil {
		b.Error("Error: Failed to read the Path !")
	}

	if typedv, _, _ := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+keys1, new(commutative.U256)); typedv == nil {
		b.Error("Error: Failed to read the key !")
	}

	u256 = new(commutative.U256).NewBoundedU256FromUint64(333, 0, 0, 999, true)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+"8975", u256); err != nil {
		b.Error(err)
	}

	typedv, _, _ := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+"8975", new(commutative.U256))
	tv := typedv.(uint256.Int)
	if tv.Cmp(uint256.NewInt(333)) != 0 {
		b.Error("Error: Failed to read the key !")
	}

	typedv, _, _ = writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", new(commutative.Path))
	if typedv == nil || typedv.(*deltaset.DeltaSet[string]).Length() != 3 {
		b.Error("Error: Failed to read the key !", typedv.(*deltaset.DeltaSet[string]).Length())
	}
}

func TestEthDataStoreAddDeleteRead(t *testing.T) {
	store := chooseDataStore()
	// store := chooseDataStore()

	// writeCache := committer.WriteCache()
	writeCache := eucache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{})

	committer := stgcommitter.NewStorageCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	committer.Sort()
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit()

	committer.Init(store)
	// create a path
	writeCache.Reset(writeCache)

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", noncommutative.NewString("path")); err == nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", new(noncommutative.Int64)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-001", noncommutative.NewInt64(2222)); err != nil {
		t.Error(err)
	}

	meta, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", new(commutative.Path))
	keys := meta.(*deltaset.DeltaSet[string]).Elements()
	if meta == nil || len(keys) != 2 ||
		keys[0] != "elem-000" ||
		keys[1] != "elem-001" {
		t.Error("not found")
	}

	// // Delete the path
	// if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", nil); err != nil {
	// 	t.Error(err)
	// }

	// if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-000", new(noncommutative.Int64)); value != nil {
	// 	t.Error("blcc://eth1.0/account/" + alice + "/storage/container/ctrn-0/elem-000 not found")
	// }

	// if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-001", new(noncommutative.Int64)); value != nil {
	// 	t.Error("blcc://eth1.0/account/" + alice + "/storage/container/ctrn-0/elem-001 not found")
	// }

	// // Write an entry having the the same name of a path, should go through
	// if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
	// 	t.Error(err)
	// }

	// if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-888", noncommutative.NewInt64(888)); err != nil {
	// 	t.Error(err)
	// }

	// if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-999", noncommutative.NewInt64(999)); err != nil {
	// 	t.Error(err)
	// }

	// if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-888", new(noncommutative.Int64)); value == nil {
	// 	t.Error("not found")
	// }

	// if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem-999", new(noncommutative.Int64)); value == nil {
	// 	t.Error("blcc://eth1.0/account/" + alice + "/storage/container/ctrn-0/elem-000 not found")
	// }

	// meta, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", &commutative.Path{})
	// keys = meta.(*deltaset.DeltaSet[string]).Elements()
	// if meta == nil || len(keys) != 2 ||
	// 	keys[0] != "elem-888" ||
	// 	keys[1] != "elem-999" {
	// 	t.Error("not found")
	// }
}
