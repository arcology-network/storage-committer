package committertest

import (
	"reflect"
	"strings"
	"testing"

	"github.com/arcology-network/common-lib/addrcompressor"
	deltaset "github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/orderedset"
	"github.com/arcology-network/common-lib/exp/slice"
	adaptorcommon "github.com/arcology-network/evm-adaptor/common"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/commutative"
	platform "github.com/arcology-network/storage-committer/platform"
	cache "github.com/arcology-network/storage-committer/storage/writecache"
	univalue "github.com/arcology-network/storage-committer/univalue"
	"github.com/holiman/uint256"
)

/* Commutative Int64 Test */
func TestTransitionFilters(t *testing.T) {
	store := chooseDataStore()

	alice := addrcompressor.RandomAccount()
	bob := addrcompressor.RandomAccount()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	// writeCache = cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache)
	// committer.NewAccount(stgcommcommon.SYSTEM, bob)

	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, bob, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := writeCache.Export(univalue.Sorter)

	acctTrans := univalue.Univalues(slice.Clone(raw)).To(univalue.IPTransition{})

	if !acctTrans[1].Value().(*commutative.U256).Equal(raw[1].Value()) {
		t.Error("Error: Non-path commutative should have the values!!")
	}

	acctTrans[0].Value().(*commutative.Path).SetSubPaths([]string{"k0", "k1"})
	acctTrans[0].Value().(*commutative.Path).SetAdded([]string{"123", "456"})
	acctTrans[0].Value().(*commutative.Path).InsertRemoved([]string{"789", "116"})

	acctTrans[1].Value().(*commutative.U256).SetValue(*uint256.NewInt(111))
	acctTrans[1].Value().(*commutative.U256).SetDelta(*uint256.NewInt(999))
	acctTrans[1].Value().(*commutative.U256).SetMin(*uint256.NewInt(1))
	acctTrans[1].Value().(*commutative.U256).SetMax(*uint256.NewInt(2222222))

	if v := raw[0].Value().(*commutative.Path).Delta().(*deltaset.DeltaSet[string]); !reflect.DeepEqual(v.Updated().Elements(), []string{}) {
		t.Error("Error: Value altered")
	}

	if v := raw[0].Value().(*commutative.Path).Delta().(*deltaset.DeltaSet[string]); !reflect.DeepEqual(v.Removed().Elements(), []string{}) {
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

	copied := univalue.Univalues(slice.Clone(acctTrans)).To(univalue.IPTransition{})

	// Test Path
	v := copied[0].Value().(*commutative.Path).Value() // Committed
	if v.(*orderedset.OrderedSet[string]).Length() != 0 {
		t.Error("Error: A path commutative variable shouldn't have the initial value")
	}

	if v := copied[0].Value().(*commutative.Path).Delta().(*deltaset.DeltaSet[string]); !reflect.DeepEqual(v.Updated().Elements(), []string{"123", "456"}) {
		t.Error("Error: Delta altered")
	}

	if v := copied[0].Value().(*commutative.Path).Delta().(*deltaset.DeltaSet[string]); !reflect.DeepEqual(v.Removed().Elements(), []string{"789", "116"}) {
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

	alice := addrcompressor.RandomAccount()
	bob := addrcompressor.RandomAccount()

	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache)
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, bob, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := writeCache.Export(univalue.Sorter)

	raw[0].Value().(*commutative.Path).SetSubPaths([]string{"k0", "k1"})
	raw[0].Value().(*commutative.Path).SetAdded([]string{"123", "456"})
	raw[0].Value().(*commutative.Path).InsertRemoved([]string{"789", "116"})

	raw[1].Value().(*commutative.U256).SetValue(*uint256.NewInt(111))
	raw[1].Value().(*commutative.U256).SetDelta(*uint256.NewInt(999))
	raw[1].Value().(*commutative.U256).SetMin(*uint256.NewInt(1))
	raw[1].Value().(*commutative.U256).SetMax(*uint256.NewInt(2222222))

	acctTrans := univalue.Univalues(slice.Clone(raw)).To(univalue.IPAccess{})

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

	idx, v := slice.FindFirstIf(acctTrans, func(v *univalue.Univalue) bool {
		return strings.Index(*v.GetPath(), "/balance") == -1 && strings.Index(*v.GetPath(), "/nonce") == -1 && v.Value() != nil
	})

	if idx != -1 {
		t.Error("Error: Nonce non-path commutative variables may keep their initial values", v)
	}
}
