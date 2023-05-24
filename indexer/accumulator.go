package indexer

import (
	"errors"

	common "github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
)

type Accumulator struct{}

func (this *Accumulator) Detect(dict *map[string]*[]ccurlcommon.UnivalueInterface) []*Conflict {
	conflicts := []*Conflict{}

	for k, transitions := range *dict {
		negatives, positives := this.Categorize(transitions)

		if underflown := this.isOutOfLimits(k, negatives); underflown != nil {
			*underflown.err = errors.New("Error: Value underflown")
			conflicts = append(conflicts, underflown)
		}

		if overflown := this.isOutOfLimits(k, positives); overflown != nil {
			*overflown.err = errors.New("Error: Value overflown")
			conflicts = append(conflicts, overflown)
		}
	}
	return conflicts
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
		err:   new(error),
	}
}

func (*Accumulator) Categorize(transitions *[]ccurlcommon.UnivalueInterface) ([]ccurlcommon.UnivalueInterface, []ccurlcommon.UnivalueInterface) {
	negatives, positives := make([]ccurlcommon.UnivalueInterface, 0), make([]ccurlcommon.UnivalueInterface, 0)
	for _, trans := range *transitions {
		if trans.Value() == nil ||
			!trans.Value().(ccurlcommon.TypeInterface).IsCommutative() ||
			!trans.Value().(ccurlcommon.TypeInterface).IsNumeric() {
			continue
		}

		common.IfThenDo(
			trans.Value().(ccurlcommon.TypeInterface).DeltaSign(),
			func() { positives = append(positives, trans) },
			func() { negatives = append(negatives, trans) })
	}

	common.RemoveIf(&negatives, func(v ccurlcommon.UnivalueInterface) bool { return v.Value() == nil })
	common.RemoveIf(&positives, func(v ccurlcommon.UnivalueInterface) bool { return v.Value() == nil })
	return negatives, positives
}
