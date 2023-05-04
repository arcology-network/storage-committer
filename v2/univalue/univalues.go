package univalue

import (
	"crypto/sha256"

	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Univalues []ccurlcommon.UnivalueInterface

func (this Univalues) To(filters ...func(ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface) []ccurlcommon.UnivalueInterface {
	for _, condition := range filters {
		this = common.CastTo(this, condition)
	}
	common.RemoveIf((*[]ccurlcommon.UnivalueInterface)(&this), func(v ccurlcommon.UnivalueInterface) bool { return v == nil })
	return this
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

func (this Univalues) Equal(other Univalues) bool {
	for i, v := range this {
		if !v.Equal(other[i]) {
			return false
		}
	}
	return true
}
