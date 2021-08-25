package concurrenturl

import (
	"bytes"
	"sort"

	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
)

type Arbimockup struct {
	transitions map[string][]ccurlcommon.UnivalueInterface
}

func (this *Arbimockup) Insert(newTrans []ccurlcommon.UnivalueInterface) {
	for _, trans := range newTrans {
		this.transitions[trans.GetPath()] = append(this.transitions[trans.GetPath()], trans)
	}
}

func (this *Arbimockup) Detect(tx uint32) []uint32 {
	for _, v := range this.transitions {
		sort.SliceStable(v, func(i, j int) bool {
			if bytes.Compare([]byte(v[i].GetPath()[:]), []byte(v[j].GetPath()[:])) == 0 {
				//	return v[i].GetTx() < v[j].GetTx()
			}
			return v[i].GetTx() < v[j].GetTx()
		})
	}
	return []uint32{}
}
