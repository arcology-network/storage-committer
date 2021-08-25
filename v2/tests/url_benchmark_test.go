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
	fmt.Println("Write:", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < 100000; i++ {
		url.Read(0, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-0"+fmt.Sprint(i)) /* The first Element */
	}
	fmt.Println("Read:", time.Since(t0))

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
	fmt.Println("map Insertion:", time.Since(t0))

	t0 = time.Now()
	m.Keys()
	fmt.Println("Get all the keys ", time.Since(t0))

	t0 = time.Now()
	targeStr := []string{}
	targeStr = GenerateStrs()
	fmt.Println("Copy keys "+fmt.Sprint(len(targeStr)), time.Since(t0))
}

func GenerateStrs() []string {
	strs := make([]string, 100000)
	for i := 0; i < 100000; i++ {
		strs[i] = "blcc://eth1.0/account/alice/storage/ctrn-0/elem-0" + fmt.Sprint(i)
	}
	return strs
}
