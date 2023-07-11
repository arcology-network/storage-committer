package indexer

import (
	"bytes"
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
		if v.Equal(target) {
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
	for i := 0; i < len(this); i++ {
		if this[i] != nil {
			ids = append(ids, this[i].GetTx())
		}
	}
	return common.UniqueInts(ids)
}

func (this Univalues) Sort(equal func(i, j int) bool, compare func(i, j int) bool) Univalues {
	// lengths := make([]int, len(this))
	// summed := make([]uint64, len(this))
	// for i := ccurlcommon.ETH10_ACCOUNT_LENGTH; i < len(this); i++ {
	// 	lengths[i] = len(*this[i].GetPath())
	// 	summed[i] = codec.String(*this[i].GetPath()).Sum(uint64(0))
	// }

	// Don't switch the order
	sort.SliceStable(this, func(i, j int) bool {
		// if lengths[i] != lengths[j] {
		// 	return lengths[i] < lengths[j]
		// }

		// if summed[i] != summed[j] {
		// 	return summed[i] < summed[j]
		// }

		if *this[i].GetPath() != *this[j].GetPath() {
			return bytes.Compare([]byte(*this[i].GetPath()), []byte(*this[j].GetPath())) < 0
		}

		if this[i].GetTx() != this[j].GetTx() {
			return this[i].GetTx() < this[j].GetTx()
		}

		if !equal(i, j) {
			return compare(i, j)
		}

		return (this[i].(*univalue.Univalue)).Less(this[j].(*univalue.Univalue))
	})
	return this
}

func (this Univalues) SortByDefault() Univalues {
	lengths := make([]int, len(this))
	summed := make([]uint64, len(this))
	for i := ccurlcommon.ETH10_ACCOUNT_LENGTH; i < len(this); i++ {
		lengths[i] = len(*this[i].GetPath())
		summed[i] = codec.String(*this[i].GetPath()).Sum(uint64(0))
	}

	// Don't switch the order
	sort.SliceStable(this, func(i, j int) bool {
		// if lengths[i] != lengths[j] {
		// 	return lengths[i] < lengths[j]
		// }

		// if summed[i] != summed[j] {
		// 	return summed[i] < summed[j]
		// }

		if *this[i].GetPath() != *this[j].GetPath() {
			return bytes.Compare([]byte(*this[i].GetPath()), []byte(*this[j].GetPath())) < 0
		}

		if this[i].GetTx() != this[j].GetTx() {
			return this[i].GetTx() < this[j].GetTx()
		}

		return true
	})
	return this
}

func (this Univalues) SortWithQuickMethod() Univalues {
	// lengths := make([]int, len(this))
	// summed := make([]uint64, len(this))
	// for i := ccurlcommon.ETH10_ACCOUNT_LENGTH; i < len(this); i++ {
	// 	// lengths[i] = len(*this[i].GetPath())
	// 	summed[i] = codec.String(*this[i].GetPath()).Sum(uint64(0))
	// }

	// Don't switch the order
	sort.SliceStable(this, func(i, j int) bool {
		// if lengths[i] != lengths[j] {
		// 	return lengths[i] < lengths[j]
		// }

		// if summed[i] != summed[j] {
		// 	return summed[i] < summed[j]
		// }

		if *this[i].GetPath() != *this[j].GetPath() {
			return bytes.Compare([]byte(*this[i].GetPath()), []byte(*this[j].GetPath())) < 0
		}

		if this[i].GetTx() != this[j].GetTx() {
			return this[i].GetTx() < this[j].GetTx()
		}

		return true
	})
	return this
}
