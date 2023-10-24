package ccurltest

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	merkle "github.com/arcology-network/common-lib/merkle"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/datatypes/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/datatypes/noncommutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	storage "github.com/arcology-network/concurrenturl/storage"
)

func TestBasicExport(t *testing.T) {
	store := ccurlcommon.NewEthMemoryDataStore()
	// store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := url.Export(indexer.Sorter)

	keys := make([]string, 0, len(acctTrans))
	values := make([][]byte, 0, len(keys))
	for i := 0; i < len(acctTrans); i++ {
		keys = append(keys, string(merkle.Sha256{}.Hash([]byte(fmt.Sprint(i))))[:3])
		values = append(values, merkle.Sha256{}.Hash([]byte(fmt.Sprint(i))))
	}

	// store.Trie().ParallelUpdate(keys, values)
	t0 := time.Now()
	store.Precommit(keys, values)
	fmt.Print(time.Since(t0))

}

func BenchmarkMultipleAccountCommitDataStore(b *testing.B) {
	store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	url := ccurl.NewConcurrentUrl(store)
	alice := AliceAccount()
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath() // create a path
	if _, err := url.Write(0, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path, true); err != nil {
		b.Error(err)
	}

	// t0 := time.Now()
	for i := 0; i < 100000; i++ {
		acct := fmt.Sprint(rand.Int())
		if err := url.NewAccount(ccurlcommon.SYSTEM, acct); err != nil { // NewAccount account structure {
			fmt.Println(err)
		}

		path := commutative.NewPath() // create a path
		if _, err := url.Write(0, "blcc://eth1.0/account/"+acct+"/storage/ctrn-0/", path, true); err != nil {
			b.Error(err)
		}

		for j := 0; j < 4; j++ {
			if _, err := url.Write(0, "blcc://eth1.0/account/"+acct+"/storage/ctrn-0/elem-0"+fmt.Sprint(j), noncommutative.NewString("fmt.Sprint(i)"), true); err != nil { /* The first Element */
				b.Error(err)
			}
		}
	}
}
