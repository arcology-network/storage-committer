package indexer

import (
	"sort"

	common "github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
)

type Accumulator struct{}

func (this *Accumulator) CheckMinMax(transitions []ccurlcommon.UnivalueInterface) []*Conflict {
	if len(transitions) <= 1 ||
		(transitions)[0].Value() == nil ||
		!(transitions)[0].Value().(ccurlcommon.TypeInterface).IsCommutative() ||
		!(transitions)[0].Value().(ccurlcommon.TypeInterface).IsNumeric() {
		return nil
	}

	common.RemoveIf(&transitions, func(v ccurlcommon.UnivalueInterface) bool {
		return v.IsReadOnly()
	})

	if len(transitions) <= 1 {
		return nil
	}

	sort.Slice(transitions, func(i, j int) bool {
		lhv := transitions[i].Value().(ccurlcommon.TypeInterface)
		rhv := transitions[i].Value().(ccurlcommon.TypeInterface)
		return lhv.DeltaSign() != rhv.DeltaSign() && !lhv.DeltaSign()
	})

	negatives, positives := this.Categorize(transitions)

	underflown := this.isOutOfLimits(*(transitions)[0].GetPath(), negatives)
	if underflown != nil {
		underflown.ErrCode = ccurlcommon.ERR_OUT_OF_LOWER_LIMIT
	}

	overflown := this.isOutOfLimits(*(transitions)[0].GetPath(), positives)
	if overflown != nil {
		overflown.ErrCode = ccurlcommon.ERR_OUT_OF_UPPER_LIMIT
	}

	if overflown == nil && underflown == nil {
		return nil
	}

	conflicts := []*Conflict{}
	if underflown != nil {
		conflicts = append(conflicts, underflown)
	}

	if overflown != nil {
		conflicts = append(conflicts, overflown)
	}
	return conflicts
}

func (*Accumulator) Categorize(transitions []ccurlcommon.UnivalueInterface) ([]ccurlcommon.UnivalueInterface, []ccurlcommon.UnivalueInterface) {
	offset := common.LocateFirstIf(transitions, func(v ccurlcommon.UnivalueInterface) bool { return v.Value().(ccurlcommon.TypeInterface).DeltaSign() })

	if offset < 0 {
		offset = len(transitions)
	}
	return transitions[:offset], transitions[offset:]
}

func (this *Accumulator) isOutOfLimits(k string, transitions []ccurlcommon.UnivalueInterface) *Conflict {
	if len(transitions) <= 1 {
		return nil
	}

	initialv := transitions[0].Value().(ccurlcommon.TypeInterface).Clone().(ccurlcommon.TypeInterface)
	_, length, err := initialv.ApplyDelta(transitions[1:])
	if err == nil {
		return nil
	}

	txIDs := []uint32{}
	common.Foreach(transitions[length+1:], func(v *ccurlcommon.UnivalueInterface) { txIDs = append(txIDs, (*v).GetTx()) })

	return &Conflict{
		key:   k,
		txIDs: txIDs,
	}
}

// func (*Accumulator) Categorize(transitions []ccurlcommon.UnivalueInterface) ([]ccurlcommon.UnivalueInterface, []ccurlcommon.UnivalueInterface) {
// 	negatives, positives := make([]ccurlcommon.UnivalueInterface, 0), make([]ccurlcommon.UnivalueInterface, 0)
// 	for _, trans := range transitions {

// 		common.IfThenDo(
// 			trans.Value().(ccurlcommon.TypeInterface).DeltaSign(),
// 			func() { positives = append(positives, trans) },
// 			func() { negatives = append(negatives, trans) })
// 	}

// 	common.RemoveIf(&negatives, func(v ccurlcommon.UnivalueInterface) bool { return v.Value() == nil })
// 	common.RemoveIf(&positives, func(v ccurlcommon.UnivalueInterface) bool { return v.Value() == nil })
// 	return negatives, positives
// }
