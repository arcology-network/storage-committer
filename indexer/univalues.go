package indexer

import (
	"crypto/sha256"
	"sort"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type Univalues []interfaces.Univalue

func (this Univalues) To(filterType interfaces.Univalue) Univalues {
	for i, v := range this {
		v := filterType.From(v)
		this[i] = common.IfThenDo1st(v != nil, func() interfaces.Univalue { return v.(interfaces.Univalue) }, nil)
	}
	common.Remove((*[]interfaces.Univalue)(&this), nil)
	return this
}

// Debugging only
func (this Univalues) IfContains(target interfaces.Univalue) bool {
	for _, v := range this {
		if (v).(interfaces.Univalue).Equal(target.(interfaces.Univalue)) {
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

func (this Univalues) UniqueTXs() []uint32 {
	var ids []uint32
	common.Foreach(this, func(v *interfaces.Univalue) { ids = append(ids, (*v).GetTx()) })
	return common.UniqueInts(ids)
}

func (this Univalues) Sort() Univalues {
	summed := make([]uint64, len(this))
	for i := ccurlcommon.ETH10_ACCOUNT_LENGTH; i < len(this); i++ {
		summed[i] = codec.String(*this[i].GetPath()).Sum(uint64(0))
	}

	// Don't switch the order
	sort.Slice(this, func(i, j int) bool {
		if summed[i] != summed[j] {
			return summed[i] < summed[j]
		}

		return (this[i].(*univalue.Univalue)).Less(this[j].(*univalue.Univalue))

		// if *this[i].GetPath() != *this[j].GetPath() {
		// 	return bytes.Compare([]byte(*this[i].GetPath()), []byte(*this[j].GetPath())) < 0
		// }

		// if this[i].GetTx() != this[j].GetTx() {
		// 	return this[i].GetTx() < this[j].GetTx()
		// }

		// if (this[i].Value() == nil || this[j].Value() == nil) && (this[i].Value() != this[j].Value()) {
		// 	return this[i].Value() == nil
		// }

		// if this[i].Writes() != this[j].Writes() {
		// 	return this[i].Writes() > this[j].Writes()
		// }

		// if this[i].Reads() != this[j].Reads() {
		// 	return this[i].Reads() > this[j].Reads()
		// }

		// if this[i].DeltaWrites() != this[j].DeltaWrites() {
		// 	return this[i].DeltaWrites() > this[j].DeltaWrites()
		// }

		// if (!this[i].Preexist() || !this[j].Preexist()) && (this[i].Preexist() != this[j].Preexist()) {
		// 	return this[i].Preexist()
		// }

		// return true
	})
	return this
}
