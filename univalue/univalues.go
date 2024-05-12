package univalue

import (
	"bytes"
	"crypto/sha256"
	"sort"
	"strings"

	"github.com/arcology-network/common-lib/exp/slice"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	intf "github.com/arcology-network/storage-committer/interfaces"
)

type Univalues []*Univalue

func (this Univalues) To(filter interface{}) Univalues {
	fun := filter.(interface{ From(*Univalue) *Univalue })
	// for i, v := range this {
	// 	this[i] = fun.From(v)
	// }

	slice.ParallelForeach(this, 8, func(i int, _ **Univalue) {
		this[i] = fun.From(this[i])
	})

	slice.Remove((*[]*Univalue)(&this), nil)
	return this
}

func (this Univalues) PathsContain(keyword string) Univalues {
	return slice.CopyIf(this, func(v *Univalue) bool {
		return strings.Contains((*v.GetPath()), (keyword))
	})
}

// Debugging only
func (this Univalues) IfContains(target *Univalue) bool {
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

func (this Univalues) Values() []intf.Type {
	vals := make([]intf.Type, len(this))
	for i, v := range this {
		vals[i] = v.Value().(intf.Type)
	}
	return vals
}

func (this Univalues) KVs() ([]string, []intf.Type) {
	keys := make([]string, len(this))
	vals := make([]intf.Type, len(this))
	for i, v := range this {
		keys[i] = *v.GetPath()
		if v.Value() == nil {
			vals[i] = nil
			continue
		}
		vals[i] = v.Value().(intf.Type)
	}
	return keys, vals
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

func (this Univalues) Clone() Univalues {
	return slice.Clone(this)
}

func (this Univalues) SortByKey() Univalues {
	sort.Slice(this, func(i, j int) bool {
		if *this[i].GetPath() == *this[j].GetPath() {
			return (*this[i].GetPath()) < (*this[j].GetPath())
		}
		return this[i].GetTx() < this[j].GetTx()
	})
	return this
}

func (this Univalues) SortByDepth() Univalues {
	depths := make([]int, len(this))
	for i, v := range this {
		depths[i] = strings.Count(*v.GetPath(), "/")
	}

	slice.SortBy1st(depths, ([]*Univalue)(this), func(i, j int) bool {
		return i < j
	})
	return this
}

func (this Univalues) Sort(groupIDs []uint32) Univalues {
	sortees := make([]struct {
		groupID uint32
		length  int
		str     string
		bytes   []byte
		tx      uint32
		value   *Univalue
	}, len(this))

	// t0 := time.Now()
	for i := 0; i < len(this); i++ {
		str := this[i].GetPath()
		bytes := []byte(*str)

		sortees[i] = struct {
			groupID uint32
			length  int
			str     string
			bytes   []byte
			tx      uint32
			value   *Univalue
		}{
			groupID: groupIDs[i],
			length:  len(bytes),
			str:     *str,
			bytes:   bytes[stgcommcommon.ETH10_ACCOUNT_PREFIX_LENGTH:],
			tx:      this[i].GetTx(),
			value:   this[i],
		}
	}

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

		if sortees[i].groupID != sortees[j].groupID {
			return sortees[i].groupID < sortees[j].groupID
		}
		return (this[i]).Less(this[j])
	}

	sort.Slice(sortees, sorter)
	for i := 0; i < len(sortees); i++ {
		this[i] = sortees[i].value
		groupIDs[i] = sortees[i].groupID
	}
	return this
}

// func (this Univalues) CompressKeys(dict *stgcommcommon.Dict) {
// 	for i, univ := range this {
// 		compressedKey := (*univ.GetPath())[stgcommcommon.ETH10_ACCOUNT_PREFIX_LENGTH:stgcommcommon.ETH10_ACCOUNT_FULL_LENGTH]
// 		newKey := dict.Compress(compressedKey, nil) + (*univ.GetPath())[stgcommcommon.ETH10_ACCOUNT_FULL_LENGTH:]
// 		this[i].SetPath(&newKey)
// 	}
// }

// func (this Univalues) DecompressKeys(dict *stgcommcommon.Dict) {
// 	for i := range this {
// 		key := *this[i].GetPath()
// 		idx := strings.Index(*this[i].GetPath(), "/")
// 		newKey := stgcommcommon.ETH10 + dict.Decompress(key[:idx]) + key[idx:]
// 		this[i].SetPath(&newKey)
// 	}
// }

func Sorter(univals []*Univalue) []*Univalue {
	sort.SliceStable(univals, func(i, j int) bool {
		lhs := (*(univals[i].GetPath()))
		rhs := (*(univals[j].GetPath()))
		return bytes.Compare([]byte(lhs)[:], []byte(rhs)[:]) < 0
	})
	return univals
}
