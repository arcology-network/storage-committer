package ccurltest

import (
	"strings"
	"testing"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	"github.com/arcology-network/common-lib/common"
	datacompression "github.com/arcology-network/common-lib/datacompression"
	ccurl "github.com/arcology-network/concurrenturl"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	univalue "github.com/arcology-network/concurrenturl/univalue"
	"github.com/holiman/uint256"
)

func TestAccumulatorUpperLimit(t *testing.T) {
	store := cachedstorage.NewDataStore()
	alice := datacompression.RandomAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	trans := univalue.Univalues(common.Clone(url.Export(univalue.Sorter))).To(univalue.TransitionCodecFilterSet()...)
	transV := []ccurlcommon.UnivalueInterface(trans)
	balanceDeltas := common.FindAll(&transV, func(v ccurlcommon.UnivalueInterface) bool { return strings.LastIndex(*v.GetPath(), "/balance") > 0 })

	balanceDeltas[0].Value().(*commutative.U256).SetMin((uint256.NewInt(0)))
	balanceDeltas[0].Value().(*commutative.U256).SetMax((uint256.NewInt(100)))
	balanceDeltas[0].Value().(*commutative.U256).SetDelta((uint256.NewInt(11)))

	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(ccurlcommon.UnivalueInterface))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(ccurlcommon.UnivalueInterface))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(ccurlcommon.UnivalueInterface))

	balanceDeltas[1].Value().(*commutative.U256).SetDelta((uint256.NewInt(21)))
	balanceDeltas[2].Value().(*commutative.U256).SetDelta((uint256.NewInt(5)))
	balanceDeltas[3].Value().(*commutative.U256).SetDelta((uint256.NewInt(63)))

	dict := make(map[string]*[]ccurlcommon.UnivalueInterface)
	dict[*(balanceDeltas[0]).GetPath()] = &balanceDeltas

	conflicts := (&indexer.Accumulator{}).Detect(&dict)
	if len(conflicts) != 0 {
		t.Error("Error: There is no conflict")
	}

	balanceDeltas[3].Value().(*commutative.U256).SetDelta((uint256.NewInt(64)))
	conflicts = (&indexer.Accumulator{}).Detect(&dict)
	if len(conflicts) != 1 {
		t.Error("Error: There is should be a of-limit-error")
	}
}

func TestAccumulatorLowerLimit(t *testing.T) {
	store := cachedstorage.NewDataStore()
	alice := datacompression.RandomAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.CreateAccount(ccurlcommon.SYSTEM, url.Platform.Eth10(), alice); err != nil { // CreateAccount account structure {
		t.Error(err)
	}

	trans := univalue.Univalues(common.Clone(url.Export(univalue.Sorter))).To(univalue.TransitionCodecFilterSet()...)
	transV := []ccurlcommon.UnivalueInterface(trans)
	balanceDeltas := common.FindAll(&transV, func(v ccurlcommon.UnivalueInterface) bool { return strings.LastIndex(*v.GetPath(), "/balance") > 0 })

	balanceDeltas[0].SetTx(0)
	balanceDeltas[0].Value().(*commutative.U256).SetMin((uint256.NewInt(0)))
	balanceDeltas[0].Value().(*commutative.U256).SetMax((uint256.NewInt(100)))
	balanceDeltas[0].Value().(*commutative.U256).SetDelta((uint256.NewInt(11)))

	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(ccurlcommon.UnivalueInterface))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(ccurlcommon.UnivalueInterface))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(ccurlcommon.UnivalueInterface))

	balanceDeltas[1].SetTx(1)
	balanceDeltas[1].Value().(*commutative.U256).SetDelta((uint256.NewInt(21)))

	balanceDeltas[2].SetTx(2)
	balanceDeltas[2].Value().(*commutative.U256).SetDelta((uint256.NewInt(5)))

	balanceDeltas[3].SetTx(3)
	balanceDeltas[3].Value().(*commutative.U256).SetDelta((uint256.NewInt(63)))

	dict := make(map[string]*[]ccurlcommon.UnivalueInterface)
	dict[*(balanceDeltas[0]).GetPath()] = &balanceDeltas

	conflicts := (&indexer.Accumulator{}).Detect(&dict)
	if len(conflicts) != 0 {
		t.Error("Error: There is no conflict")
	}

	balanceDeltas[3].Value().(*commutative.U256).SetDelta((uint256.NewInt(64)))
	conflicts = (&indexer.Accumulator{}).Detect(&dict)
	if len(conflicts) != 1 {
		t.Error("Error: There is should be a of-limit-error")
	}
}
