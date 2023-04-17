package ccurltype

import (
	"bytes"
	"crypto/sha256"
	"sort"

	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Univalues []ccurlcommon.UnivalueInterface

func (this Univalues) Sort() {
	sort.SliceStable(this, func(i, j int) bool {
		lhs := (*(this[i].GetPath()))
		rhs := (*(this[j].GetPath()))
		return bytes.Compare([]byte(lhs)[:], []byte(rhs)[:]) < 0
	})
	//return this
}

func (this Univalues) IfContains(condition ccurlcommon.UnivalueInterface) bool {
	for _, v := range this {
		if (v).(*Univalue).Equal(condition.(*Univalue)) {
			return true
		}
	}
	return false
}

func (this Univalues) Keys() []string {
	keys := make([]string, len(this))
	for i, v := range this {
		keys[i] = *v.GetPath()
	}
	return keys
}

// For debug only
func (this Univalues) Checksum() [32]byte {
	return sha256.Sum256(this.Encode())
}

// Only work when users aren't allowed to give arbitrary names
// func (this Univalues) SetTransitType(trans []ccurlcommon.UnivalueInterface, func(name string) ()) {
// 	for i := 0; i < len(trans); i++ {
// 		if strings.Contains(*(trans[i].GetPath()), name) {
// 			trans[i].SetTransitionType(ccurlcommon.INVARIATE_TRANSITIONS)
// 		}
// 	}
// }
