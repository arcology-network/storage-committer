package ccurltest

import (
	"strings"
	"testing"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurl "github.com/arcology-network/concurrenturl"
	arbitrator "github.com/arcology-network/concurrenturl/arbitrator"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/interfaces"
	storage "github.com/arcology-network/concurrenturl/storage"
)

func TestAccumulatorUpperLimit(t *testing.T) {
	store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)

	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	trans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	transV := []interfaces.Univalue(trans)
	balanceDeltas := common.CopyIf(transV, func(v interfaces.Univalue) bool { return strings.LastIndex(*v.GetPath(), "/balance") > 0 })

	// v := (&codec.Uint256{}).NewInt(0)
	balanceDeltas[0].Value().(*commutative.U256).SetMin((&codec.Uint256{}).NewInt(0))
	balanceDeltas[0].Value().(*commutative.U256).SetMax((&codec.Uint256{}).NewInt(100))
	balanceDeltas[0].Value().(*commutative.U256).SetDelta((&codec.Uint256{}).NewInt(11))

	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(interfaces.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(interfaces.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(interfaces.Univalue))

	balanceDeltas[1].Value().(*commutative.U256).SetDelta((&codec.Uint256{}).NewInt(21))
	balanceDeltas[2].Value().(*commutative.U256).SetDelta((&codec.Uint256{}).NewInt(5))
	balanceDeltas[3].Value().(*commutative.U256).SetDelta((&codec.Uint256{}).NewInt(63))

	// dict := make(map[string]*[]interfaces.Univalue)
	// dict[*(balanceDeltas[0]).GetPath()] = &balanceDeltas

	conflicts := (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if len(conflicts) != 0 {
		t.Error("Error: There is no conflict")
	}

	balanceDeltas[3].Value().(*commutative.U256).SetDelta((&codec.Uint256{}).NewInt(64))
	conflicts = (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if len(conflicts) != 1 {
		t.Error("Error: There is should be a of-limit-error")
	}
}

func TestAccumulatorLowerLimit(t *testing.T) {
	store := cachedstorage.NewDataStore(nil, nil, nil, storage.Codec{}.Encode, storage.Codec{}.Decode)
	alice := AliceAccount()
	url := ccurl.NewConcurrentUrl(store)
	if err := url.NewAccount(ccurlcommon.SYSTEM, alice); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	trans := indexer.Univalues(common.Clone(url.Export(indexer.Sorter))).To(indexer.ITCTransition{})
	transV := []interfaces.Univalue(trans)
	balanceDeltas := common.CopyIf(transV, func(v interfaces.Univalue) bool { return strings.LastIndex(*v.GetPath(), "/balance") > 0 })

	balanceDeltas[0].SetTx(0)
	balanceDeltas[0].Value().(*commutative.U256).SetMin(((&codec.Uint256{}).NewInt(0)))
	balanceDeltas[0].Value().(*commutative.U256).SetMax(((&codec.Uint256{}).NewInt(100)))
	balanceDeltas[0].Value().(*commutative.U256).SetDelta(((&codec.Uint256{}).NewInt(11)))

	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(interfaces.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(interfaces.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(interfaces.Univalue))

	balanceDeltas[1].SetTx(1)
	balanceDeltas[1].Value().(*commutative.U256).SetDelta(((&codec.Uint256{}).NewInt(21)))

	balanceDeltas[2].SetTx(2)
	balanceDeltas[2].Value().(*commutative.U256).SetDelta(((&codec.Uint256{}).NewInt(5)))

	balanceDeltas[3].SetTx(3)
	balanceDeltas[3].Value().(*commutative.U256).SetDelta(((&codec.Uint256{}).NewInt(63)))

	conflicts := (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if len(conflicts) != 0 {
		t.Error("Error: There is no conflict")
	}

	balanceDeltas[3].Value().(*commutative.U256).SetDelta(((&codec.Uint256{}).NewInt(64)))
	conflicts = (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if len(conflicts) != 1 {
		t.Error("Error: There is should be a of-limit-error")
	}
}
