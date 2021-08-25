package ccurltest

import (
	"fmt"
	"reflect"
	"testing"

	ccurl "github.com/HPISTechnologies/concurrenturl/v2"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
	commutative "github.com/HPISTechnologies/concurrenturl/v2/type/commutative"
	noncommutative "github.com/HPISTechnologies/concurrenturl/v2/type/noncommutative"
)

func TestMetaIterator(t *testing.T) {
	store := ccurlcommon.NewDataStore()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.Preload(ccurlcommon.SYSTEM, url.Platform.Eth10(), "alice"); err != nil { // Preload account structure {
		t.Error(err)
	}

	_, acctTrans := url.Export(false)
	url.Commit(acctTrans, []uint32{ccurlcommon.SYSTEM})

	path, _ := commutative.NewMeta("blcc://eth1.0/account/alice/storage/ctrn-0/")
	url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/", path)

	for i := 0; i < 5; i++ {
		url.Write(1, "blcc://eth1.0/account/alice/storage/ctrn-0/elem-"+fmt.Sprint(i), noncommutative.NewInt64(int64(i)))
	}

	/* Forward Iter */
	v, _ := url.Read(1, "blcc://eth1.0/account/alice/storage/ctrn-0/")
	target := []string{
		"elem-0",
		"elem-1",
		"elem-2",
		"elem-3",
		"elem-4",
	}

	results := []string{}
	for i := 0; i < 5; i++ {
		results = append(results, v.(*commutative.Meta).Next())
	}

	if !reflect.DeepEqual(results, target) {
		t.Error("Error: Wrong iterator values")
	}

	results = []string{}
	v.(*commutative.Meta).ResetIterator()
	for i := 0; i < 5; i++ {
		results = append(results, v.(*commutative.Meta).Next())
	}

	if !reflect.DeepEqual(results, target) {
		t.Error("Error: Wrong iterator values")
	}

	rTarget := []string{
		"elem-4",
		"elem-3",
		"elem-2",
		"elem-1",
		"elem-0",
	}

	results = []string{}
	for i := 0; i < 5; i++ {
		results = append(results, v.(*commutative.Meta).Previous())
	}

	if !reflect.DeepEqual(results, rTarget) {
		t.Error("Error: Wrong reverse iterator values")
	}

	results = []string{}
	v.(*commutative.Meta).ResetReverseIterator()
	for i := 0; i < 5; i++ {
		results = append(results, v.(*commutative.Meta).Previous())
	}

	if !reflect.DeepEqual(results, rTarget) {
		t.Error("Error: Wrong reverse iterator values")
	}

}
