package committertest

import (
	"strings"
	"testing"

	"github.com/arcology-network/common-lib/exp/slice"
	adaptorcommon "github.com/arcology-network/evm-adaptor/common"
	arbitrator "github.com/arcology-network/storage-committer/arbitrator"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/commutative"
	platform "github.com/arcology-network/storage-committer/platform"
	cache "github.com/arcology-network/storage-committer/storage/writecache"
	univalue "github.com/arcology-network/storage-committer/univalue"
	"github.com/holiman/uint256"
)

func TestAccumulatorUpperLimit(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	itc := univalue.ITTransition{}
	trans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(itc)
	transV := []*univalue.Univalue(trans)
	balanceDeltas := slice.CopyIf(transV, func(v *univalue.Univalue) bool { return strings.LastIndex(*v.GetPath(), "/balance") > 0 })

	// v := *uint256.NewInt(0)
	balanceDeltas[0].Value().(*commutative.U256).SetMin(*uint256.NewInt(0))
	balanceDeltas[0].Value().(*commutative.U256).SetMax(*uint256.NewInt(100))
	balanceDeltas[0].Value().(*commutative.U256).SetDelta(*uint256.NewInt(11))

	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*univalue.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*univalue.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*univalue.Univalue))

	balanceDeltas[1].Value().(*commutative.U256).SetDelta(*uint256.NewInt(21))
	balanceDeltas[2].Value().(*commutative.U256).SetDelta(*uint256.NewInt(5))
	balanceDeltas[3].Value().(*commutative.U256).SetDelta(*uint256.NewInt(63))

	// dict := make(map[string]*[]*univalue.Univalue)
	// dict[*(balanceDeltas[0]).GetPath()] = &balanceDeltas

	conflicts := (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if len(conflicts) != 0 {
		t.Error("Error: There is no conflict")
	}

	balanceDeltas[3].Value().(*commutative.U256).SetDelta(*uint256.NewInt(64))
	conflicts = (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if len(conflicts) != 1 {
		t.Error("Error: There is should be a of-limit-error")
	}
}

func TestAccumulatorLowerLimit(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	// url := stgcommitter.NewStateCommitter(store)
	writeCache := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	trans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	transV := []*univalue.Univalue(trans)
	balanceDeltas := slice.CopyIf(transV, func(v *univalue.Univalue) bool { return strings.LastIndex(*v.GetPath(), "/balance") > 0 })

	balanceDeltas[0].SetTx(0)
	balanceDeltas[0].Value().(*commutative.U256).SetMin((*uint256.NewInt(0)))
	balanceDeltas[0].Value().(*commutative.U256).SetMax((*uint256.NewInt(100)))
	balanceDeltas[0].Value().(*commutative.U256).SetDelta((*uint256.NewInt(11)))

	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*univalue.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*univalue.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*univalue.Univalue))

	balanceDeltas[1].SetTx(1)
	balanceDeltas[1].Value().(*commutative.U256).SetDelta((*uint256.NewInt(21)))

	balanceDeltas[2].SetTx(2)
	balanceDeltas[2].Value().(*commutative.U256).SetDelta((*uint256.NewInt(5)))

	balanceDeltas[3].SetTx(3)
	balanceDeltas[3].Value().(*commutative.U256).SetDelta((*uint256.NewInt(63)))

	conflicts := (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if len(conflicts) != 0 {
		t.Error("Error: There is no conflict")
	}

	balanceDeltas[3].Value().(*commutative.U256).SetDelta((*uint256.NewInt(64)))
	conflicts = (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if len(conflicts) != 1 {
		t.Error("Error: There is should be a of-limit-error")
	}
}
