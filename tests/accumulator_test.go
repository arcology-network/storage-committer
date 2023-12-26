package ccurltest

import (
	"strings"
	"testing"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl"
	ccurl "github.com/arcology-network/concurrenturl"
	arbitrator "github.com/arcology-network/concurrenturl/arbitrator"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/interfaces"
	"github.com/holiman/uint256"
)

func TestAccumulatorUpperLimit(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	writeCache := url.WriteCache()

	if _, err := concurrenturl.CreateNewAccount(ccurlcommon.SYSTEM, alice, ccurlcommon.NewPlatform(), writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	trans := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	transV := []interfaces.Univalue(trans)
	balanceDeltas := common.CopyIf(transV, func(v interfaces.Univalue) bool { return strings.LastIndex(*v.GetPath(), "/balance") > 0 })

	// v := *uint256.NewInt(0)
	balanceDeltas[0].Value().(*commutative.U256).SetMin(*uint256.NewInt(0))
	balanceDeltas[0].Value().(*commutative.U256).SetMax(*uint256.NewInt(100))
	balanceDeltas[0].Value().(*commutative.U256).SetDelta(*uint256.NewInt(11))

	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(interfaces.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(interfaces.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(interfaces.Univalue))

	balanceDeltas[1].Value().(*commutative.U256).SetDelta(*uint256.NewInt(21))
	balanceDeltas[2].Value().(*commutative.U256).SetDelta(*uint256.NewInt(5))
	balanceDeltas[3].Value().(*commutative.U256).SetDelta(*uint256.NewInt(63))

	// dict := make(map[string]*[]interfaces.Univalue)
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
	url := ccurl.NewConcurrentUrl(store)
	writeCache := url.WriteCache()
	if _, err := concurrenturl.CreateNewAccount(ccurlcommon.SYSTEM, alice, ccurlcommon.NewPlatform(), writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	trans := indexer.Univalues(common.Clone(writeCache.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	transV := []interfaces.Univalue(trans)
	balanceDeltas := common.CopyIf(transV, func(v interfaces.Univalue) bool { return strings.LastIndex(*v.GetPath(), "/balance") > 0 })

	balanceDeltas[0].SetTx(0)
	balanceDeltas[0].Value().(*commutative.U256).SetMin((*uint256.NewInt(0)))
	balanceDeltas[0].Value().(*commutative.U256).SetMax((*uint256.NewInt(100)))
	balanceDeltas[0].Value().(*commutative.U256).SetDelta((*uint256.NewInt(11)))

	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(interfaces.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(interfaces.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(interfaces.Univalue))

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
