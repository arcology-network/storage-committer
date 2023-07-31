package ccurltest

import (
	"testing"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/datacompression"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

func TestItcIsolation(t *testing.T) {
	// compressionLut := datacompression.NewCompressionLut()
	store := cachedstorage.NewDataStore()
	alice := datacompression.RandomAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	original := common.Clone(url.Export(indexer.Sorter))
	cloned := indexer.Univalues(original).Clone()
	acctTrans := indexer.Univalues(cloned).To(indexer.IPCTransition{})

	for i := 0; i < len(acctTrans); i++ {
		meta := acctTrans[i].GetUnimeta().(univalue.Unimeta)
		(meta).IncrementReads(1)
		(meta).IncrementWrites(2)
		(meta).IncrementDeltaWrites(3)
		v := acctTrans[i].Value().(interfaces.Type)
		v.ResetDelta()
		// NewPathDelta(add []string, del []string)
	}

	if !cloned.Equal(original) {
		t.Error("Value altered")
	}

	// buffer := indexer.Univalues(acctTrans).Encode()
	// out := indexer.Univalues{}.Decode(buffer).(indexer.Univalues)

	// if _, err := url.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil, true); err != nil { // Delete the path
	// 	t.Error(err)
	// }
}
