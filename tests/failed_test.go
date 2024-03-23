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

	commoncache "github.com/arcology-network/common-lib/cache"
	"github.com/arcology-network/common-lib/exp/slice"
	cache "github.com/arcology-network/eu/cache"
	stgcommitter "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/commutative"
	importer "github.com/arcology-network/storage-committer/importer"
	"github.com/arcology-network/storage-committer/interfaces"

	noncommutative "github.com/arcology-network/storage-committer/noncommutative"
	platform "github.com/arcology-network/storage-committer/platform"
	storage "github.com/arcology-network/storage-committer/storage"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

func TestAddAndDelete(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore(nil, nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)
	committer := stgcommitter.NewStorageCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	alice := AliceAccount()
	if _, err := writeCache.CreateNewAccount(stgcommcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)

	// acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import(univalue.Univalues{})

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(0)

	committer.SetStore(store)
	// path := commutative.NewPath()
	// writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/c", path)

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", noncommutative.NewString("124")); err != nil {
		t.Error(err)
	}

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	committer = stgcommitter.NewStorageCommitter(store)
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{1})
	committer.Commit(0)

	committer.SetStore(store)
	writeCache.Reset()

	// Delete an non-existing entry, should NOT appear in the transitions
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/4", nil); err == nil {
		t.Error("Deleting an non-existing entry should've flaged an error", err)
	}

	raw := univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.ITTransition{})
	if acctTrans := raw; len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}

	// Delete an non-existing entry, should NOT appear in the transitions
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/4", noncommutative.NewString("124")); err != nil {
		t.Error("Failed to write", err)
	}

	if v, _, err := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/4", nil); v != "124" {
		t.Error("Wrong return value", err)
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", new(noncommutative.String)); v != "124" {
		t.Error("Error: Wrong return value")
	}
}

func TestPathReadAndWriteBatchCache(b *testing.T) {
	store := chooseDataStore()
	store.(*storage.StoreRouter).EnableCache()
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

	committer := stgcommitter.NewStorageCommitter(store).Import(univalue.Univalues(slice.Clone(writeCache.Export(importer.Sorter))).To(importer.IPTransition{}))
	committer.Precommit([]uint32{0})
	committer.Commit(0)

	for i := 0; i < len(keys); i++ {
		v, ok := store.(*storage.StoreRouter).Cache(nil).(*commoncache.ReadCache[string, interfaces.Type]).Get("blcc://eth1.0/account/" + alice + "/storage/container/ctrn-0/alice-elem-" + keys[i])
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
	committer = stgcommitter.NewStorageCommitter(store).Import(trans)
	committer.Precommit([]uint32{0})
	committer.Commit(0)

	for i := 0; i < len(keys); i++ {
		v, ok := store.(*storage.StoreRouter).Cache(nil).(*commoncache.ReadCache[string, interfaces.Type]).Get("blcc://eth1.0/account/" + alice + "/storage/container/ctrn-0/alice-elem-" + keys[i])
		if typedv, _, _ := (*(v)).Get(); !ok || typedv != int64(i+9999) {
			b.Error("not found")
		}
	}
}
