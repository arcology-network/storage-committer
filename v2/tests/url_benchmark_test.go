package ccurltest

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	merkle "github.com/HPISTechnologies/common-lib/merkle"
	ccurl "github.com/HPISTechnologies/concurrenturl/v2"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
	commutative "github.com/HPISTechnologies/concurrenturl/v2/type/commutative"
	noncommutative "github.com/HPISTechnologies/concurrenturl/v2/type/noncommutative"
	orderedmap "github.com/elliotchance/orderedmap"
)

func BenchmarkSingleAccountCommit(b *testing.B) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	if err := url.Write(0, "blcc://eth1.0/account/alice/storage/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	t0 := time.Now()
	for i := 0; i < 100000; i++ {
		if err := url.Write(0, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-0"+fmt.Sprint(i), noncommutative.NewString("fmt.Sprint(i)")); err != nil { /* The first Element */
			b.Error(err)
		}
	}
	fmt.Println("First Write:", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < 100000; i++ {
		url.Read(0, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-0"+fmt.Sprint(i)) /* The first Element */
	}
	fmt.Println("First Read:", time.Since(t0))

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

func BenchmarkMultipleAccountCommit(b *testing.B) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	if err := url.Write(0, "blcc://eth1.0/account/alice/storage/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	t0 := time.Now()
	for i := 0; i < 2500; i++ {
		acct := fmt.Sprint(rand.Int())
		if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), acct); err != nil { // Preload account structure {
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
	url.Commit(trans, []uint32{ccurlcommon.SYSTEM})

	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
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
	url.Commit(trans, []uint32{ccurlcommon.SYSTEM})

	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
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
	url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice")
	_, acctTrans := url.Export(false)
	url.Commit(acctTrans, []uint32{ccurlcommon.SYSTEM})

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
