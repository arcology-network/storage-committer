package concurrenturl

import (
	"fmt"
	"testing"
	"time"

	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/type/noncommutative"
)

func BenchmarkWriteRead(b *testing.B) {
	store := ccurlcommon.NewDataStore()
	url := NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), "Alice"); err != nil { // CreateAccount account structure {
		fmt.Println(err)
	}

	if err := url.Initialize(url.Platform.Eth10(), "Alice"); err != nil { // Initialize account paths
		fmt.Println(err)
	}

	path, _ := commutative.NewMeta("blcc://eth1.0/Alice/storage/ctrn-0/") // create a path
	if err := url.Write(0, "blcc://eth1.0/Alice/storage/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	t0 := time.Now()
	for i := 0; i < 100000; i++ {
		if err := url.Write(0, "blcc://eth1.0/Alice/storage/ctrn-0/elem-0"+fmt.Sprint(i), noncommutative.NewString("fmt.Sprint(i)")); err != nil { /* The first Element */
			b.Error(err)
		}
	}
	fmt.Println("Write:", time.Since(t0))

	t0 = time.Now()
	for i := 0; i < 100000; i++ {
		url.Read(0, "blcc://eth1.0/Alice/storage/ctrn-0/elem-0"+fmt.Sprint(i)) /* The first Element */
	}
	fmt.Println("Read:", time.Since(t0))

	t0 = time.Now()
	_, trans := url.Export()
	fmt.Println("Export:", time.Since(t0))

	t0 = time.Now()
	url.indexer.Import(trans)
	fmt.Println("Import:", time.Since(t0))

	t0 = time.Now()
	err := url.indexer.Commit([]uint32{0, 1})
	if err != nil {
		return
	}
	fmt.Println("Commit:", time.Since(t0))
}
