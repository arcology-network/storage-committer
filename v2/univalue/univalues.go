package univalue

import (
	"crypto/sha256"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Univalues []ccurlcommon.UnivalueInterface

func (this Univalues) Filter(processor func(ccurlcommon.UnivalueInterface) codec.Encodeable) []codec.Encodeable {
	encodeables := common.CastTo(this, processor)
	common.RemoveIf(&encodeables, func(v codec.Encodeable) bool { return v == nil })
	return encodeables
}

func CastTo[T0, T1 any](src []T0, predicate func(T0) T1) []T1 {
	target := make([]T1, len(src))
	for i := range src {
		target[i] = predicate(src[i])
	}
	return target
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
