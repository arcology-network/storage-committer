package ccurltest

import (
	"reflect"
	"strings"
	"testing"

	datacompression "github.com/arcology-network/common-lib/addrcompressor"
	"github.com/arcology-network/common-lib/common"
	orderedset "github.com/arcology-network/common-lib/container/set"
	"github.com/arcology-network/concurrenturl"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/interfaces"
	"github.com/holiman/uint256"
)

/* Commutative Int64 Test */
func TestTransitionFilters(t *testing.T) {
	store := chooseDataStore()

	alice := datacompression.RandomAccount()
	bob := datacompression.RandomAccount()

	url := ccurl.NewConcurrentUrl(store)
	writeCache := url.WriteCache()
	concurrenturl.CreateNewAccount(ccurlcommon.SYSTEM, alice, ccurlcommon.NewPlatform(), writeCache)
	// url.NewAccount(ccurlcommon.SYSTEM, bob)

	if _, err := concurrenturl.CreateNewAccount(ccurlcommon.SYSTEM, bob, ccurlcommon.NewPlatform(), writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := url.WriteCache().Export(indexer.Sorter)

	acctTrans := indexer.Univalues(common.Clone(raw)).To(indexer.IPCTransition{})

	if !acctTrans[1].Value().(*commutative.U256).Equal(raw[1].Value()) {
		t.Error("Error: Non-path commutative should have the values!!")
	}

	acctTrans[0].Value().(*commutative.Path).SetSubs([]string{"k0", "k1"})
	acctTrans[0].Value().(*commutative.Path).SetAdded([]string{"123", "456"})
	acctTrans[0].Value().(*commutative.Path).SetRemoved([]string{"789", "116"})

	acctTrans[1].Value().(*commutative.U256).SetValue(*uint256.NewInt(111))
	acctTrans[1].Value().(*commutative.U256).SetDelta(*uint256.NewInt(999))
	acctTrans[1].Value().(*commutative.U256).SetMin(*uint256.NewInt(1))
	acctTrans[1].Value().(*commutative.U256).SetMax(*uint256.NewInt(2222222))

	if v := raw[0].Value().(*commutative.Path).Delta().(*commutative.PathDelta); !reflect.DeepEqual(v.Added(), []string{}) {
		t.Error("Error: Value altered")
	}

	if v := raw[0].Value().(*commutative.Path).Delta().(*commutative.PathDelta); !reflect.DeepEqual(v.Removed(), []string{}) {
		t.Error("Error: Delta altered")
	}

	if v := raw[1].Value().(*commutative.U256).Delta().(uint256.Int); !v.Eq(uint256.NewInt(0)) {
		t.Error("Error: Value altered")
	}

	if v := raw[1].Value().(*commutative.U256).Delta().(uint256.Int); !v.Eq(uint256.NewInt(0)) {
		t.Error("Error: Delta altered")
	}

	if v := raw[1].Value().(*commutative.U256).Min().(uint256.Int); !v.Eq(&commutative.U256_MIN) {
		t.Error("Error: Min Value altered")
	}

	if v := raw[1].Value().(*commutative.U256).Max().(uint256.Int); !v.Eq(&commutative.U256_MAX) {
		t.Error("Error: Max altered")
	}

	copied := indexer.Univalues(common.Clone(acctTrans)).To(indexer.IPCTransition{})

	// Test Path
	if v := copied[0].Value().(*commutative.Path).Value().(*orderedset.OrderedSet); len(v.Keys()) != 0 {
		t.Error("Error: A path commutative variable shouldn't have the initial value")
	}

	if v := copied[0].Value().(*commutative.Path).Delta().(*commutative.PathDelta); !reflect.DeepEqual(v.Added(), []string{"123", "456"}) {
		t.Error("Error: Delta altered")
	}

	if v := copied[0].Value().(*commutative.Path).Delta().(*commutative.PathDelta); !reflect.DeepEqual(v.Removed(), []string{"789", "116"}) {
		t.Error("Error: Delta altered")
	}

	// Test Commutative 256
	if v := copied[1].Value().(*commutative.U256).Value().(uint256.Int); !(&v).Eq(uint256.NewInt(111)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}
	if v := copied[1].Value().(*commutative.U256).Delta().(uint256.Int); !(&v).Eq(uint256.NewInt(999)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

	if v := copied[1].Value().(*commutative.U256).Min().(uint256.Int); !(&v).Eq(uint256.NewInt(1)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

	if v := copied[1].Value().(*commutative.U256).Max().(uint256.Int); !(&v).Eq(uint256.NewInt(2222222)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

}

func TestAccessFilters(t *testing.T) {
	store := chooseDataStore()

	alice := datacompression.RandomAccount()
	bob := datacompression.RandomAccount()

	url := ccurl.NewConcurrentUrl(store)
	writeCache := url.WriteCache()
	concurrenturl.CreateNewAccount(ccurlcommon.SYSTEM, alice, ccurlcommon.NewPlatform(), writeCache)
	if _, err := concurrenturl.CreateNewAccount(ccurlcommon.SYSTEM, bob, ccurlcommon.NewPlatform(), writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := url.WriteCache().Export(indexer.Sorter)

	raw[0].Value().(*commutative.Path).SetSubs([]string{"k0", "k1"})
	raw[0].Value().(*commutative.Path).SetAdded([]string{"123", "456"})
	raw[0].Value().(*commutative.Path).SetRemoved([]string{"789", "116"})

	raw[1].Value().(*commutative.U256).SetValue(*uint256.NewInt(111))
	raw[1].Value().(*commutative.U256).SetDelta(*uint256.NewInt(999))
	raw[1].Value().(*commutative.U256).SetMin(*uint256.NewInt(1))
	raw[1].Value().(*commutative.U256).SetMax(*uint256.NewInt(2222222))

	acctTrans := indexer.Univalues(common.Clone(raw)).To(indexer.IPCAccess{})

	if acctTrans[0].Value() != nil {
		t.Error("Error: Value altered")
	}

	// Test Commutative 256
	if v := acctTrans[1].Value().(*commutative.U256).Value().(uint256.Int); !(&v).Eq(uint256.NewInt(111)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}
	if v := acctTrans[1].Value().(*commutative.U256).Delta().(uint256.Int); !(&v).Eq(uint256.NewInt(999)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

	if v := acctTrans[1].Value().(*commutative.U256).Min().(uint256.Int); !(&v).Eq(uint256.NewInt(1)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

	if v := acctTrans[1].Value().(*commutative.U256).Max().(uint256.Int); !(&v).Eq(uint256.NewInt(2222222)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

	idx, v := common.FindFirstIf(acctTrans, func(v interfaces.Univalue) bool {
		return strings.Index(*v.GetPath(), "/balance") == -1 && strings.Index(*v.GetPath(), "/nonce") == -1 && v.Value() != nil
	})

	if idx != -1 {
		t.Error("Error: Nonce non-path commutative variables may keep their initial values", v)
	}

}
