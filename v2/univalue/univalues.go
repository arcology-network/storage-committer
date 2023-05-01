package univalue

import (
	"crypto/sha256"

	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Univalues []ccurlcommon.UnivalueInterface

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

func (this Univalues) Equal(other Univalues) bool {
	for i, v := range this {
		if !v.Equal(other[i]) {
			return false
		}
	}
	return true
}
