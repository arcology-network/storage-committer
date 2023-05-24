package indexer

import (
	common "github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
)

type Conflict struct {
	key   string
	txIDs []uint32
	err   *error
}

type Conflicts []Conflict

func (this Conflicts) IDs() []uint32 {
	txIDs := []uint32{}
	for _, v := range this {
		txIDs = append(txIDs, v.txIDs...)
	}
	return txIDs
}

func (this Conflicts) Keys() []string {
	keys := []string{}
	for _, v := range this {
		keys = append(keys, v.key)
	}
	return keys
}

type Accumulator struct{}

func (this *Accumulator) Detect(dict *map[string]*[]ccurlcommon.UnivalueInterface) []Conflict {
	conflicts := []Conflict{}

	for k, transitions := range *dict {
		negatives, positives := this.Categorize(transitions)
		underflown := this.isOutOfLimits(k, negatives)
		overflown := this.isOutOfLimits(k, positives)

		if underflown != nil {
			conflicts = append(conflicts, *underflown)
		}

		if overflown != nil {
			conflicts = append(conflicts, *overflown)
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

	txIDs := []uint32{}
	common.Foreach(transitions[:length], func(v *ccurlcommon.UnivalueInterface) { txIDs = append(txIDs, (*v).GetTx()) })

	return &Conflict{
		txIDs: []uint32{},
		err:   &err,
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

		if trans.Value().(ccurlcommon.TypeInterface).Sign() {
			positives = append(positives, trans)
		} else {
			negatives = append(negatives, trans)
		}

		common.IfThenDo(
			trans.Value().(ccurlcommon.TypeInterface).Sign(),
			func() { positives = append(positives, trans) },
			func() { negatives = append(negatives, trans) })
	}

	common.RemoveIf(&negatives, func(v ccurlcommon.UnivalueInterface) bool { return v.Value() == nil })
	common.RemoveIf(&positives, func(v ccurlcommon.UnivalueInterface) bool { return v.Value() == nil })
	return negatives, positives
}
