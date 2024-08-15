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
	"math/rand"
	"sort"
	"testing"
	"time"

	addrcompressor "github.com/arcology-network/common-lib/addrcompressor"
	"github.com/arcology-network/common-lib/exp/slice"
	stgcommcommon "github.com/arcology-network/common-lib/types/storage/common"
	commutative "github.com/arcology-network/common-lib/types/storage/commutative"
	noncommutative "github.com/arcology-network/common-lib/types/storage/noncommutative"
	univalue "github.com/arcology-network/common-lib/types/storage/univalue"
	adaptorcommon "github.com/arcology-network/evm-adaptor/common"
	statestore "github.com/arcology-network/storage-committer"
	stgcommitter "github.com/arcology-network/storage-committer/storage/committer"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	orderedmap "github.com/elliotchance/orderedmap"
	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/btree"
	"golang.org/x/crypto/sha3"
	// "github.com/google/btree"
	// ehtrlp "github.com/elliotchance/orderedmap"
)

func BenchmarkAccountMerkleImportPerf(b *testing.B) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	writeCache := sstore.WriteCache
	for i := 0; i < 1000; i++ {
		if _, err := adaptorcommon.CreateNewAccount(0, fmt.Sprint(rand.Float64()), writeCache); err != nil { // Preload account structure {
			b.Error(err)
		}
	}
	acct := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITAccess{})

	t0 := time.Now()
	univalue.Univalues(acct).Encode()
	b.Log("Transition Encoding: ", len(acct), time.Since(t0))
}

func BenchmarkSingleAccountCommit(b *testing.B) {
	store := chooseDataStore()

	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	writeCache := sstore.WriteCache
	alice := AliceAccount()
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath() // create a path
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	for i := 0; i < 1; i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0"+fmt.Sprint(i), noncommutative.NewString("fmt.Sprint(i)")); err != nil { /* The first Element */
			b.Error(err)
		}
	}

	transitions := univalue.Univalues(slice.Clone(writeCache.Export())).To(univalue.ITTransition{})

	t0 := time.Now()
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(transitions)
	committer.Precommit([]uint32{stgcommcommon.SYSTEM, 0, 1})
	committer.Commit(10)
	fmt.Println("Init committer= :", time.Since(t0), "with initial transitions:", len(transitions))
}

func BenchmarkMultipleAccountCommit(b *testing.B) {
	store := chooseDataStore()

	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	writeCache := sstore.WriteCache
	alice := AliceAccount()
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath() // create a path
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	t0 := time.Now()
	numAccounts := 25000
	for i := 0; i < numAccounts; i++ {
		buf := sha3.Sum256([]byte(fmt.Sprint(i)))
		acct := hexutil.Encode(buf[:20])

		// writeCache := committer.WriteCache()
		if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, acct, writeCache); err != nil { // NewAccount account structure {
			fmt.Println(err)
		}

		path := commutative.NewPath() // create a path
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+acct+"/storage/ctrn-0/", path); err != nil {
			b.Error(err)
		}

		for j := 0; j < 40; j++ {
			if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0"+fmt.Sprint(j), noncommutative.NewString(string(acct))); err != nil { /* The first Element */
				b.Error(err)
			}
		}
	}
	fmt.Println("Created ", (numAccounts), "Accounts in:", time.Since(t0))

	t0 = time.Now()
	trans := slice.Clone(writeCache.Export())
	fmt.Println("Clone:", len(trans), "in ", time.Since(t0))

	t0 = time.Now()
	trans = univalue.Univalues(trans).To(univalue.IPTransition{})
	fmt.Println("To(univalue.ITTransition{}):", len(trans), "in ", time.Since(t0))

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	t0 = time.Now()
	committer.Import(trans)
	fmt.Println("Import: ", len(trans), " in: ", time.Since(t0))

	t0 = time.Now()

	fmt.Println("Sort: ", len(trans), " in: ", time.Since(t0))

	t0 = time.Now()
	committer.Precommit([]uint32{0})
	fmt.Println("Precommit:", time.Since(t0))

	fmt.Println("Root: ", sstore.Backend().EthStore().LatestWorldTrieRoot())

	t0 = time.Now()
	committer.Commit(10)
	fmt.Println("Commit:", time.Since(t0))

	t0 = time.Now()
	// nilHash := merkle.Sha256(nil)
	// fmt.Println("Hash:", nilHash)
	fmt.Println("merkle: ", time.Since(t0))

	t0 = time.Now()
	writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0"+fmt.Sprint(0), new(noncommutative.String))

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/"+fmt.Sprint(0), noncommutative.NewString(string("acct"))); err != nil { /* The first Element */
		b.Error(err)
	}
	fmt.Println("Write: 2", time.Since(t0))
}

func BenchmarkAddThenDelete(b *testing.B) {
	// store := chooseDataStore()
	store := chooseDataStore()

	// sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))

	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	writeCache := sstore.WriteCache
	meta := commutative.NewPath()
	writeCache.Write(stgcommcommon.SYSTEM, stgcommcommon.ETH10_ACCOUNT_PREFIX, meta)
	trans := univalue.Univalues(slice.Clone(writeCache.Export())).To(univalue.ITTransition{})

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(trans)

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(10)

	alice := AliceAccount()
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	t0 := time.Now()
	for i := 0; i < 50000; i++ {
		_, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-"+fmt.Sprint(i), noncommutative.NewInt64(int64(i)))
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Write "+fmt.Sprint(50000), time.Since(t0))

	t0 = time.Now()
	for i := 0; i < 50000; i++ {
		if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-"+fmt.Sprint(i), nil); err != nil {
			panic(err)
		}
	}
	fmt.Println("Deleted 50000 keys "+fmt.Sprint(50000), time.Since(t0))
}

func BenchmarkAddThenPop(b *testing.B) {
	// store := chooseDataStore()

	store := chooseDataStore()

	// writeCache := committer.WriteCache()
	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	writeCache := sstore.WriteCache
	meta := commutative.NewPath()
	writeCache.Write(stgcommcommon.SYSTEM, stgcommcommon.ETH10_ACCOUNT_PREFIX, meta)

	trans := univalue.Univalues(slice.Clone(writeCache.Export())).To(univalue.ITTransition{})

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(10)

	alice := AliceAccount()
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	t0 := time.Now()
	for i := 0; i < 50000; i++ {
		v := noncommutative.NewBytes([]byte(fmt.Sprint(rand.Float64())))
		_, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-"+fmt.Sprint(i), v)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Write "+fmt.Sprint(50000), "noncommutative bytes in", time.Since(t0))
}

// func BenchmarkOrderedMap(b *testing.B) {
// 	m := orderedmap.NewOrderedMap()
// 	alice := AliceAccount()
// 	t0 := time.Now()

// 	for i := 0; i < 100000; i++ {
// 		m.Set("blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0"+fmt.Sprint(i))
// 	}
// 	fmt.Println("orderedmap Insertion:", time.Since(t0))

// 	t0 = time.Now()
// 	m2 := make(map[string]bool)
// 	for i := 0; i < 100000; i++ {
// 		m2["blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0"+fmt.Sprint(i)] = true
// 	}
// 	fmt.Println("Golang map Insertion:", time.Since(t0))

// 	t0 = time.Now()
// 	m.Keys()
// 	fmt.Println("orderedmap get keys ", time.Since(t0))

// 	t0 = time.Now()
// 	targeStr := make([]string, 100000)
// 	for i := 0; i < 100000; i++ {
// 		targeStr[i] = "blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-0" + fmt.Sprint(i)
// 	}
// 	fmt.Println("Copy keys "+fmt.Sprint(len(targeStr)), time.Since(t0))
// }

// func BenchmarkOrderedMapInit(b *testing.B) {
// 	t0 := time.Now()
// 	orderedMaps := make([]*orderedmap.OrderedMap, 100000)
// 	for i := 0; i < len(orderedMaps); i++ {
// 		orderedMaps[i] = orderedmap.NewOrderedMap()
// 	}
// 	fmt.Println("Initialized  "+fmt.Sprint(len(orderedMaps)), "OrderedMap in", time.Since(t0))
// }

func BenchmarkInsertAndDelete(b *testing.B) {
	m := orderedmap.NewOrderedMap()
	alice := AliceAccount()

	t0 := time.Now()
	for i := 0; i < 100000; i++ {
		m.Set("blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0"+fmt.Sprint(i), 0)
	}
	fmt.Println("orderedmap Insertion:", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < 100000; i++ {
		//	m.Delete("blcc://eth1.0/account/" + alice +"/storage/ctrn-0/elem-0" + fmt.Sprint(i))
		m.Delete("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-0" + fmt.Sprint(i))
	}

	fmt.Println("Delete then delete 100000 keys:", time.Since(t0))
}

func BenchmarkMapInit(b *testing.B) {
	t0 := time.Now()
	orderedMaps := make([]map[string]bool, 100000)
	for i := 0; i < len(orderedMaps); i++ {
		orderedMaps[i] = make(map[string]bool)
	}
	fmt.Println("Initialized  "+fmt.Sprint(len(orderedMaps)), "OrderedMap in", time.Since(t0))
}

func BenchmarkShrinkSlice(b *testing.B) {
	strs := make([]string, 100000)
	alice := AliceAccount()
	for i := 0; i < len(strs); i++ {
		strs[i] = "blcc://eth1.0/account/" + alice + "/storage/ctrn-0/elem-0" + fmt.Sprint(i)
	}

	t0 := time.Now()
	for i := 0; i < len(strs)/10; i++ {
		idx := rand.Int() % (len(strs) - 2)
		copy(strs[idx:len(strs)-2], strs[idx+1:len(strs)-1])
	}
	fmt.Println("Remove random element from a Slice ", "from", 100000, "to", len(strs), "in", time.Since(t0))
}

func BenchmarkEncodeTransitions(b *testing.B) {
	store := chooseDataStore()

	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	writeCache := sstore.WriteCache

	alice := AliceAccount()
	adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache)
	// acctTrans := univalue.Univalues(slice.Clone(writeCache.Export())).To(univalue.ITAccess{})

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITAccess{})

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(10)

	path := commutative.NewPath()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	t0 := time.Now()
	for i := 0; i < 100000; i++ {
		writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-"+fmt.Sprint(i), noncommutative.NewInt64(int64(i)))
	}
	fmt.Println("Write "+fmt.Sprint(10000), time.Since(t0))

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITAccess{})

	t0 = time.Now()
	univalue.Univalues(acctTrans).Encode()
	fmt.Println("Encode "+fmt.Sprint(len(acctTrans)), time.Since(t0))

	/* Forward Iter */
	// t0 = time.Now()
	// v, _ := committer.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	// for i := 0; i < 100000; i++ {
	// 	v.(*commutative.Path).Next()
	// }
	// fmt.Println("Next "+fmt.Sprint(100000), time.Since(t0))

	// v.(*commutative.Path).ResetIterator()
	// for i := 0; i < 100000; i++ {
	// 	v.(*commutative.Path).Next()
	// }

	// for i := 0; i < 100000; i++ {
	// 	v.(*commutative.Path).Previous()
	// }

	// v.(*commutative.Path).ResetReverseIterator()
	// for i := 0; i < 100000; i++ {
	// 	v.(*commutative.Path).Previous()
	// }
}

func BenchmarkAccountCreationWithMerkle(b *testing.B) {
	// lut := addrcompressor.NewCompressionLut()
	// fileDB, err := datastore.NewFileDB(ROOT_PATH, 8, 2)
	// if err != nil {
	// 	b.Error(err)
	// 	return
	// }
	store := chooseDataStore()
	// store := datastore.NewDataStore( nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)
	// store.Inject((stgcommcommon.ETH10_ACCOUNT_PREFIX), commutative.NewPath())

	t0 := time.Now()

	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	writeCache := sstore.WriteCache
	for i := 0; i < 10; i++ {
		acct := addrcompressor.RandomAccount()
		if _, err := adaptorcommon.CreateNewAccount(0, acct, writeCache); err != nil { // Preload account structure {
			b.Error(err)
		}
	}
	fmt.Println("Write "+fmt.Sprint(100000*9), time.Since(t0))

	t0 = time.Now()
	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export())).To(univalue.ITTransition{})

	fmt.Println("Export "+fmt.Sprint(100000*9), time.Since(t0))

	t0 = time.Now()

	// transitions := univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues)
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans)

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(10)
	// errs := committer.AllInOneCommit(acctTrans, []uint32{0})

	// if len(errs) > 0 {
	// 	fmt.Println(errs)
	// }
	fmt.Println("Commit + Merkle "+fmt.Sprint(100000*9), time.Since(t0))
}

// func TestOrderedMapBasic(t *testing.T) {
// 	om := orderedmap.NewOrderedMap()
// 	om.Set("abc", 1)
// 	om.Set("xyz", 2)
// 	om.Set("uvw", 3)
// 	om.Set("def", 4)
// 	for iter := om.Front(); iter != nil; iter = iter.Next() {
// 		t.Log(iter.Key, iter.Value)
// 	}
// }

// func TestLLRB(t *testing.T) {
// 	tree := llrb.New()

// 	tree.ReplaceOrInsert(llrb.String("abc"))
// 	tree.ReplaceOrInsert(llrb.String("xyz"))
// 	tree.ReplaceOrInsert(llrb.String("uvw"))
// 	tree.ReplaceOrInsert(llrb.String("def"))

// 	tree.AscendGreaterOrEqual(tree.Min(), func(i llrb.Item) bool {
// 		t.Log(i)
// 		return true
// 	})
// }

// func TestPathRepeats(t *testing.T) {
// 	paths := make([]string, 0, 2)
// 	for i := 0; i < 1; i++ {
// 		acct := addrcompressor.RandomAccount()
// 		for j := 0; j < 10; j++ {
// 			paths = append(paths, (&stgcommitter.Platform{}).Eth10Account()+acct+"/"+fmt.Sprint(rand.Float64()))
// 		}
// 	}

// 	positions := make([]int, 0, len(paths))
// 	positions = append(positions, 0)
// 	current := paths[0]
// 	for i := 1; i < len(paths); i++ {
// 		p0 := current[:len((&stgcommitter.Platform{}).Eth10Account())+stgcommcommon.ETH10_ACCOUNT_LENGTH]
// 		p1 := paths[i][:len((&stgcommitter.Platform{}).Eth10Account())+stgcommcommon.ETH10_ACCOUNT_LENGTH]
// 		if p0 != p1 {
// 			current = paths[i]
// 			positions = append(positions, i)
// 		}
// 	}
// 	positions = append(positions, len(paths))
// }

func BenchmarkStringSort(b *testing.B) {
	paths := make([][]*univalue.Univalue, 100000)
	for i := 0; i < 100000; i++ {
		acct := addrcompressor.RandomAccount()
		for j := 9; j >= 1; j-- {

			paths[i] = append(paths[i], univalue.NewUnivalue(uint32(j), acct, 0, 0, 0, noncommutative.NewString(fmt.Sprint(rand.Float64())), nil))
		}
	}

	t0 := time.Now()

	slice.ParallelForeach(paths, 6, func(i int, _ *[]*univalue.Univalue) {
		sort.SliceStable(paths[i], func(i, j int) bool {
			if paths[i][j].GetTx() == stgcommcommon.SYSTEM {
				return true
			}

			if paths[i][j].GetTx() == stgcommcommon.SYSTEM {
				return false
			}

			return paths[i][j].GetTx() < paths[i][j].GetTx()
		})
	})

	fmt.Println("Path Sort "+fmt.Sprint(100000*9), time.Since(t0))
}

type String string

func (s String) Less(b btree.Item) bool {
	return s < b.(String)
}

// func BenchmarkOrderedMapPerf(b *testing.B) {
// 	N := 1000000
// 	ss := make([]string, N)
// 	for i := 0; i < N; i++ {
// 		ss[i] = "blcc://eth1.0/account/storage/containers/" + fmt.Sprint(rand.Float64())
// 	}

// 	t0 := time.Now()
// 	gomap := make(map[string]string)
// 	for i := 0; i < N; i++ {
// 		gomap[ss[i]] = ss[i]
// 	}
// 	b.Log("time of go map set:", time.Since(t0))

// 	t0 = time.Now()
// 	tlen0 := 0
// 	for i := 0; i < N; i++ {
// 		tlen0 += len(gomap[ss[i]])
// 	}
// 	b.Log("time of go map get:", time.Since(t0))

// 	t0 = time.Now()
// 	omap := orderedmap.NewOrderedMap()
// 	for i := 0; i < N; i++ {
// 		omap.Set(ss[i], ss[i])
// 	}
// 	b.Log("time of orderedmap set:", time.Since(t0))

// 	t0 = time.Now()
// 	tlen1 := 0
// 	for iter := omap.Front(); iter != nil; iter = iter.Next() {
// 		tlen1 += len(iter.Value.(string))
// 	}
// 	b.Log("time of orderedmap get:", time.Since(t0))

// 	t0 = time.Now()
// 	tree := llrb.New()
// 	for i := 0; i < N; i++ {
// 		tree.ReplaceOrInsert(llrb.String(ss[i]))
// 	}
// 	b.Log("time of llrb insert:", time.Since(t0))

// 	t0 = time.Now()
// 	tlen2 := 0
// 	tree.AscendGreaterOrEqual(tree.Min(), func(i llrb.Item) bool {
// 		tlen2 += len(i.(llrb.String))
// 		return true
// 	})
// 	b.Log("time of llrb get:", time.Since(t0))

// 	t0 = time.Now()
// 	btr := btree.New(32)
// 	for i := 0; i < N; i++ {
// 		btr.ReplaceOrInsert(String(ss[i]))
// 	}
// 	b.Log("time of btree insert:", time.Since(t0))

// 	t0 = time.Now()
// 	tlen3 := 0
// 	btr.AscendGreaterOrEqual(btr.Min(), func(i btree.Item) bool {
// 		tlen3 += len(i.(String))
// 		return true
// 	})
// 	b.Log("time of btree get:", time.Since(t0))

// 	t0 = time.Now()
// 	sort.Strings(ss)
// 	b.Log("time of go sort:", time.Since(t0))

// 	if tlen0 != tlen1 || tlen0 != tlen2 {
// 		b.Fail()
// 	}
// }

// func TestHashPerformance(t *testing.T) {
// 	h1 := fnv1a.HashString64("Hello World!")
// 	fmt.Println("FNV-1a hash of 'Hello World!':", h1)

// 	records := make([]string, 10000)
// 	for i := 0; i < len(records); i++ {
// 		records[i] = (&stgcommitter.Platform{}).Eth10() + addrcompressor.RandomAccount()
// 	}

// 	t0 := time.Now()
// 	for i := 0; i < len(records); i++ {
// 		h0, h1 := murmur.Sum128(codec.String(records[i]).Encode())
// 		records[i] = (codec.Bytes(codec.Uint64(h0).Encode()).ToString() + codec.Bytes(codec.Uint64(h1).Encode()).ToString())
// 	}
// 	fmt.Println("murmur "+fmt.Sprint(10000), time.Since(t0))

// 	t0 = time.Now()
// 	for i := 0; i < len(records); i++ {
// 		h0 := fnv1a.HashString64(records[i])
// 		records[i] = codec.Bytes(codec.Uint64(h0).Encode()).ToString() + codec.Bytes(codec.Uint64(h0).Encode()).ToString()

// 	}
// 	fmt.Println("fnv1a "+fmt.Sprint(10000), time.Since(t0))

// 	hash, _ := murmur.Sum128([]byte("FNV-1a hash of 'Hello World!':"))
// 	fmt.Println(hash)
// }

// func BenchmarkTransitionImport(b *testing.B) {
// 	store := chooseDataStore()
// 	meta := commutative.NewPath()
// 	store.Inject((stgcommcommon.ETH10_ACCOUNT_PREFIX), meta)

// 	t0 := time.Now()

// 	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
// 	writeCache := sstore.WriteCache

// 	// writeCache := committer.WriteCache()
// 	for i := 0; i < 150000; i++ {
// 		acct := addrcompressor.RandomAccount()
// 		if _, err := adaptorcommon.CreateNewAccount(0, acct, writeCache); err != nil { // Preload account structure {
// 			b.Error(err)
// 		}
// 	}
// 	fmt.Println("Write "+fmt.Sprint(100000*9), time.Since(t0))

// 	t0 = time.Now()
// 	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export())).To(univalue.ITAccess{})

// 	fmt.Println("Export "+fmt.Sprint(150000*9), time.Since(t0))

// 	// accountMerkle := importer.NewAccountMerkle(platform.NewPlatform(), rlpEncoder, merkle.Keccak256{}.Hash)

// 	fmt.Println("-------------")
// 	t0 = time.Now()
// 	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
// 	committer.Import(acctTrans)
// 	// accountMerkle.Import(acctTrans)
// 	fmt.Println("committer + accountMerkle Import "+fmt.Sprint(150000*9), time.Since(t0))
// }

// func BenchmarkConcurrentTransitionImport(b *testing.B) {
// 	store := datastore.NewDataStore( nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)
// 	meta := commutative.NewPath()
// 	store.Inject((stgcommcommon.ETH10_ACCOUNT_PREFIX), meta)

// 	t0 := time.Now()
// 		committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
// writeCache := committer.WriteCache()
// 	for i := 0; i < 90000; i++ {
// 		acct := addrcompressor.RandomAccount()
// 		if _, err := adaptorcommon.CreateNewAccount(0, acct);err != nil { // Preload account structure {
// 			b.Error(err)
// 		}
// 	}
// 	fmt.Println("Write "+fmt.Sprint(100000*9), time.Since(t0))

// 	t0 = time.Now()
// 	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export())).To(univalue.ITAccess{})

// 	fmt.Println("Export "+fmt.Sprint(150000*9), time.Since(t0))

// 	accountMerkle := importer.NewAccountMerkle(platform.NewPlatform(), rlpEncoder, merkle.Keccak256{}.Hash)

// 	t0 = time.Now()
// 	common.ParallelExecute(
// 		func() { committer.Import(acctTrans) },
// 		func() { accountMerkle.Import(acctTrans) },
// 	)
// 	fmt.Println("ParallelExecute Import "+fmt.Sprint(150000*9), time.Since(t0))
// }

func BenchmarkRandomAccountSort(t *testing.B) {
	// store := chooseDataStore()
	// meta := commutative.NewPath()
	// store.Inject((stgcommcommon.ETH10_ACCOUNT_PREFIX), meta)

	// t0 := time.Now()

	// sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	// writeCache := sstore.WriteCache
	// for i := 0; i < 100000; i++ {
	// 	acct := addrcompressor.RandomAccount()
	// 	if _, err := adaptorcommon.CreateNewAccount(0, acct, writeCache); err != nil { // Preload account structure {
	// 		// b.Error(err)
	// 	}
	// }
	// fmt.Println("Write "+fmt.Sprint(100000*9), time.Since(t0))

	// t0 = time.Now()
	// in := univalue.Univalues(slice.Clone(writeCache.Export())).To(univalue.ITAccess{})

	// t0 = time.Now()
	// univalue.Univalues(in).Sort(nil)
	// fmt.Println("Univalues(in).Sort()", len(in), "entires in :", time.Since(t0))

	// t0 = time.Now()
	// univalue.Univalues(in).SortByDefault()
	// fmt.Println("Univalues(in).SortByDefault()", len(in), "entires in :", time.Since(t0))

	// t0 = time.Now()
	// univalue.Univalues(in).SortWithQuickMethod()
	// fmt.Println("Univalues(in).SortWithQuickMethod()", len(in), "entires in :", time.Since(t0))

}
