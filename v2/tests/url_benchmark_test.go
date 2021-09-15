package ccurltest

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"testing"
	"time"

	merkle "github.com/arcology-network/common-lib/merkle"
	ccurl "github.com/arcology-network/concurrenturl/v2"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	ccurltype "github.com/arcology-network/concurrenturl/v2/type"
	commutative "github.com/arcology-network/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/v2/type/noncommutative"
	orderedmap "github.com/elliotchance/orderedmap"
)

func BenchmarkSingleAccountCommit(b *testing.B) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // CreateAccount account structure {
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	for i := 0; i < 100000; i++ {
		if err := url.Write(0, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-0"+fmt.Sprint(i), noncommutative.NewString("fmt.Sprint(i)")); err != nil { /* The first Element */
			b.Error(err)
		}
	}

	//t0 = time.Now()
	_, transitions := url.Export(false)
	in := ccurltype.Univalues(transitions).Encode()
	out := ccurltype.Univalues{}.Decode(in).(ccurltype.Univalues)

	//fmt.Println("Export:", time.Since(t0))

	t0 := time.Now()
	url.Import(out)
	url.Commit([]uint32{0})
	fmt.Println("Total :", time.Since(t0))
}

func BenchmarkMultipleAccountCommit(b *testing.B) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // CreateAccount account structure {
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	if err := url.Write(0, "blcc://eth1.0/account/alice/storage/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	t0 := time.Now()
	for i := 0; i < 2500; i++ {
		acct := fmt.Sprint(rand.Int())
		if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), acct); err != nil { // CreateAccount account structure {
			fmt.Println(err)
		}

		path, _ := commutative.NewMeta("blcc://eth1.0/account/" + acct + "/storage/ctrn-0/") // create a path
		if err := url.Write(0, "blcc://eth1.0/account/"+acct+"/storage/ctrn-0/", path); err != nil {
			b.Error(err)
		}

		for j := 0; j < 4; j++ {
			if err := url.Write(0, "blcc://eth1.0/account/"+acct+"/storage/ctrn-0/elem-0"+fmt.Sprint(j), noncommutative.NewString("fmt.Sprint(i)")); err != nil { /* The first Element */
				b.Error(err)
			}
		}
	}
	fmt.Println("Write:", time.Since(t0))

	t0 = time.Now()
	_, trans := url.Export(false)
	fmt.Println("Export:", time.Since(t0))

	t0 = time.Now()
	url.Indexer().Import(trans)
	fmt.Println("Import:", time.Since(t0))

	t0 = time.Now()
	_, _, err := url.Indexer().Commit([]uint32{0})
	fmt.Println("Commit:", time.Since(t0))

	if len(err) > 0 {
		b.Error(err)
	}

	nilHash := merkle.Sha256(nil)
	fmt.Print(nilHash)
}

func BenchmarkUrlAddThenDelete(b *testing.B) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	meta, _ := commutative.NewMeta(ccurlcommon.NewPlatform().Eth10Account())
	url.Write(ccurlcommon.SYSTEM, ccurlcommon.NewPlatform().Eth10Account(), meta)
	_, trans := url.Export(false)
	url.Import(trans)
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // CreateAccount account structure {
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/")
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", path)

	t0 := time.Now()
	for i := 0; i < 50000; i++ {
		err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-"+fmt.Sprint(i), noncommutative.NewInt64(int64(i)))
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Write "+fmt.Sprint(50000), time.Since(t0))

	t0 = time.Now()
	for i := 0; i < 50000; i++ {
		if err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-"+fmt.Sprint(i), nil); err != nil {
			panic(err)
		}
	}
	fmt.Println("Deleted 50000 keys "+fmt.Sprint(50000), time.Since(t0))
}

func BenchmarkUrlAddThenPop(b *testing.B) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	meta, _ := commutative.NewMeta(ccurlcommon.NewPlatform().Eth10Account())
	url.Write(ccurlcommon.SYSTEM, ccurlcommon.NewPlatform().Eth10Account(), meta)
	_, trans := url.Export(false)
	url.Import(trans)
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // CreateAccount account structure {
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/")
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", path)

	t0 := time.Now()
	for i := 0; i < 50000; i++ {
		v := noncommutative.NewBytes([]byte(fmt.Sprint(rand.Float64())))
		err := url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-"+fmt.Sprint(i), v)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Write "+fmt.Sprint(50000), "noncommutative bytes in", time.Since(t0))

	t0 = time.Now()
	v, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	for i := 0; i < 50000; i++ {
		key := v.(*commutative.Meta).Next()
		url.Write(1, key, nil)
	}
	fmt.Println("Pop 50000 noncommutative bytes in", fmt.Sprint(50000), time.Since(t0))
}

func BenchmarkOrderedMap(b *testing.B) {
	m := orderedmap.NewOrderedMap()

	t0 := time.Now()
	for i := 0; i < 100000; i++ {
		m.Set("blcc://eth1.0/account/alice/storage/ctrn-0/elem-0"+fmt.Sprint(i), true)
	}
	fmt.Println("orderedmap Insertion:", time.Since(t0))

	t0 = time.Now()
	m2 := make(map[string]bool)
	for i := 0; i < 100000; i++ {
		m2["blcc://eth1.0/account/alice/storage/ctrn-0/elem-0"+fmt.Sprint(i)] = true
	}
	fmt.Println("Golang map Insertion:", time.Since(t0))

	t0 = time.Now()
	m.Keys()
	fmt.Println("orderedmap get keys ", time.Since(t0))

	t0 = time.Now()
	targeStr := make([]string, 100000)
	for i := 0; i < 100000; i++ {
		targeStr[i] = "blcc://eth1.0/account/alice/storage/ctrn-0/elem-0" + fmt.Sprint(i)
	}
	fmt.Println("Copy keys "+fmt.Sprint(len(targeStr)), time.Since(t0))
}

func BenchmarkOrderedMapInit(b *testing.B) {
	t0 := time.Now()
	orderedMaps := make([]*orderedmap.OrderedMap, 100000)
	for i := 0; i < len(orderedMaps); i++ {
		orderedMaps[i] = orderedmap.NewOrderedMap()
	}
	fmt.Println("Initialized  "+fmt.Sprint(len(orderedMaps)), "OrderedMap in", time.Since(t0))
}

func BenchmarkInsertAndDelete(b *testing.B) {
	m := orderedmap.NewOrderedMap()
	t0 := time.Now()
	for i := 0; i < 100000; i++ {
		m.Set("blcc://eth1.0/account/alice/storage/ctrn-0/elem-0"+fmt.Sprint(i), true)
	}
	fmt.Println("orderedmap Insertion:", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < 100000; i++ {
		//	m.Delete("blcc://eth1.0/account/alice/storage/ctrn-0/elem-0" + fmt.Sprint(i))
		m.Delete("blcc://eth1.0/account/alice/storage/ctrn-0/elem-0" + fmt.Sprint(i))
	}

	fmt.Println("Delete then delete keys:", time.Since(t0))
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
	for i := 0; i < len(strs); i++ {
		strs[i] = "blcc://eth1.0/account/alice/storage/ctrn-0/elem-0" + fmt.Sprint(i)
	}

	// t0 := time.Now()
	// for i := 0; i < len(strs); i++ {
	// 	strs = strs[:len(strs)-i]
	// }
	// fmt.Println("Back shrink Slice ", "from", 100000, "to", len(strs), "in", time.Since(t0))

	t0 := time.Now()
	for i := 0; i < len(strs)/10; i++ {
		idx := rand.Int() % (len(strs) - 2)
		copy(strs[idx:len(strs)-2], strs[idx+1:len(strs)-1])
	}
	fmt.Println("Remove random element from a Slice ", "from", 100000, "to", len(strs), "in", time.Since(t0))
}

func BenchmarkMetaIterator(b *testing.B) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice")
	_, acctTrans := url.Export(false)
	url.Import(acctTrans)
	url.Commit([]uint32{ccurlcommon.SYSTEM})

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/")
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", path)

	t0 := time.Now()
	for i := 0; i < 100000; i++ {
		url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-"+fmt.Sprint(i), noncommutative.NewInt64(int64(i)))
	}
	fmt.Println("Write "+fmt.Sprint(10000), time.Since(t0))

	/* Forward Iter */
	t0 = time.Now()
	v, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	for i := 0; i < 100000; i++ {
		v.(*commutative.Meta).Next()
	}
	fmt.Println("Next "+fmt.Sprint(100000), time.Since(t0))

	v.(*commutative.Meta).ResetIterator()
	for i := 0; i < 100000; i++ {
		v.(*commutative.Meta).Next()
	}

	for i := 0; i < 100000; i++ {
		v.(*commutative.Meta).Previous()
	}

	v.(*commutative.Meta).ResetReverseIterator()
	for i := 0; i < 100000; i++ {
		v.(*commutative.Meta).Previous()
	}
}

func BenchmarkMapKeyLengthComparison(b *testing.B) {
	t0 := time.Now()
	short := make([]string, 100000)
	long := make([]string, 100000)
	for i := 0; i < 100000; i++ {
		long[i] = "blcc://eth1.0/account/alice/storage/ctrn-0/elem-" + fmt.Sprint(i)
		short[i] = fmt.Sprint(i)
	}
	fmt.Println("Write "+fmt.Sprint(100000), time.Since(t0))

	t0 = time.Now()
	longMap := make(map[[32]byte]string)
	for i := 0; i < len(long); i++ {
		longMap[sha256.Sum256([]byte(long[i]))] = long[i]
	}
	fmt.Println("longMap / [32]byte key"+fmt.Sprint(100000), time.Since(t0))

	t0 = time.Now()
	shortKeyMap := make(map[string]string)
	for i := 0; i < len(short); i++ {
		shortKeyMap[short[i]] = long[i]
	}
	fmt.Println("shortMap / short key"+fmt.Sprint(100000), time.Since(t0))

	t0 = time.Now()
	longkeyshortMap := make(map[string]string)
	for i := 0; i < len(short); i++ {
		longkeyshortMap[short[i]] = long[i]
	}
	fmt.Println("shortMap / long key"+fmt.Sprint(100000), time.Since(t0))
}

func BenchmarkAccountCreationWithMerkle(b *testing.B) {
	store := ccurlcommon.NewDataStore()
	meta, _ := commutative.NewMeta((ccurlcommon.NewPlatform().Eth10Account()))
	store.Save((ccurlcommon.NewPlatform().Eth10Account()), meta)

	t0 := time.Now()
	url := ccurl.NewConcurrentUrl(store)
	for i := 0; i < 100000; i++ {
		if err := url.CreateAccount(0, (url.Platform.Eth10()), fmt.Sprint(rand.Float64())); err != nil { // Preload account structure {
			b.Error(err)
		}
	}
	fmt.Println("Write "+fmt.Sprint(100000*9), time.Since(t0))

	t0 = time.Now()
	_, acctTrans := url.Export(false)
	fmt.Println("Export "+fmt.Sprint(100000*9), time.Since(t0))

	t0 = time.Now()
	//url.Import(acctTrans)
	errs := url.AllInOneCommit(acctTrans, []uint32{0})
	if len(errs) > 0 {
		fmt.Println(errs)
	}
	fmt.Println("Commit + Merkle "+fmt.Sprint(100000*9), time.Since(t0))
}

func BenchmarkMapKeyForSharding(b *testing.B) {
	total := 0
	//keys := make([]string, 100000)
	key := fmt.Sprint(rand.Int())
	t0 := time.Now()
	for i := 0; i < 10; i++ {
		for j := 0; j < len(key); j++ {
			//fmt.Print(key[j])
			total += int(key[j])
		}
	}
	fmt.Println("Total :", total, fmt.Sprint(100000*9), time.Since(t0))
}

// func TestStringEngine(t *testing.T) {
// 	paths := make([]string, 4)
// 	paths[2] = "000000000000000000"
// 	paths[1] = "1111111111111111111110"
// 	paths[0] = "2222222222222222"
// 	paths[3] = "2222222222222222"

// 	se := mhasher.Start()
// 	err := se.ToBuffer(paths)
// 	if err != nil {
// 		fmt.Printf("ToBuffer err: %v\n", err)
// 		return
// 	}

// 	retpaths, err := se.FromBuffer(paths)
// 	if err != nil {
// 		fmt.Printf("FromBuffer err: %v\n", err)
// 		return
// 	}

// 	for i := range retpaths {
// 		fmt.Printf("paths=%x\n", retpaths[i])
// 	}

// 	se.Clear()
// 	se.Stop()
// }
