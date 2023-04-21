package ccurltest

import (
	"testing"
)

func TestMetaIterator(t *testing.T) {
	// store := cachedstorage.NewDataStore()
	// url := ccurl.NewConcurrentUrl(store)
	// alice := datacompression.RandomAccount()
	// if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
	// 	t.Error(err)
	// }

	// _, acctTrans := url.Export(false)
	// url.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	// url.PostImport()
	// url.Commit([]uint32{ccurlcommon.SYSTEM})

	// path, _ := commutative.NewMeta("blcc://eth1.0/account/" + alice + "/storage/ctrn-0/")
	// url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path)

	// for i := 0; i < 5; i++ {
	// 	url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-"+fmt.Sprint(i), noncommutative.NewInt64(int64(i)))
	// }

	// /* Forward Iter */
	// v, _ := url.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/")
	// results := []string{}
	// for i := 0; i < 5; i++ {
	// 	results = append(results, v.(*commutative.Meta).Next())
	// }

	// target := []string{
	// 	"elem-0",
	// 	"elem-1",
	// 	"elem-2",
	// 	"elem-3",
	// 	"elem-4",
	// }

	// if !reflect.DeepEqual(results, target) {
	// 	t.Error("Error: Wrong iterator values !")
	// }

	// results = []string{}
	// v.(*commutative.Meta).ResetIterator()
	// for i := 0; i < 5; i++ {
	// 	results = append(results, v.(*commutative.Meta).Next())
	// }

	// if !reflect.DeepEqual(results, target) {
	// 	t.Error("Error: Wrong iterator values")
	// }

	// rTarget := []string{
	// 	"elem-4",
	// 	"elem-3",
	// 	"elem-2",
	// 	"elem-1",
	// 	"elem-0",
	// }

	// results = []string{}
	// for i := 0; i < 5; i++ {
	// 	results = append(results, v.(*commutative.Meta).Previous())
	// }

	// if !reflect.DeepEqual(results, rTarget) {
	// 	t.Error("Error: Wrong reverse iterator values")
	// }

	// results = []string{}
	// v.(*commutative.Meta).ResetReverseIterator()
	// for i := 0; i < 5; i++ {
	// 	results = append(results, v.(*commutative.Meta).Previous())
	// }

	// if !reflect.DeepEqual(results, rTarget) {
	// 	t.Error("Error: Wrong reverse iterator values")
	// }

}
