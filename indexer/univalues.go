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
	sortees := make([]struct {
		length int
		str    string
		bytes  []byte
		tx     uint32
		value  interfaces.Univalue
	}, len(this))

	// t0 := time.Now()
	for i := 0; i < len(this); i++ {
		str := this[i].GetPath()
		bytes := *codec.UnsafeStringToBytes(str) // 100% faster than ([]byte(*str))

		sortees[i] = struct {
			length int
			str    string
			bytes  []byte
			tx     uint32
			value  interfaces.Univalue
		}{
			length: len(bytes),
			str:    *str,
			bytes:  bytes[ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH:],
			tx:     this[i].GetTx(),
			value:  this[i],
		}
	}
	// fmt.Println("Sorted ", len(this), "entires in :", time.Since(t0), "Total size: ")

	sorter := func(i, j int) bool {
		if sortees[i].length != sortees[j].length {
			return sortees[i].length < sortees[j].length
		}

		if sortees[i].str != sortees[j].str {
			return bytes.Compare(sortees[i].bytes, sortees[j].bytes) < 0
		}

		if sortees[i].tx != sortees[j].tx {
			return sortees[i].tx < sortees[j].tx
		}

		if equal != nil && compare != nil && !equal(i, j) {
			return compare(i, j)
		}

		return (this[i].(*univalue.Univalue)).Less(this[j].(*univalue.Univalue))
	}

	sort.Slice(sortees, sorter)

	for i := 0; i < len(sortees); i++ {
		this[i] = sortees[i].value
	}
	return this
}

func (this Univalues) SortByDefault() Univalues {
	// lengths := make([]int, len(this))
	// summed := make([]uint64, len(this))
	// for i := ccurlcommon.ETH10_ACCOUNT_LENGTH; i < len(this); i++ {
	// 	lengths[i] = len(*this[i].GetPath())
	// 	summed[i] = codec.String(*this[i].GetPath()).Sum(uint64(0))
	// }

	// Don't switch the order
	sort.Slice(this, func(i, j int) bool {
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
	sort.Slice(this, func(i, j int) bool {
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
