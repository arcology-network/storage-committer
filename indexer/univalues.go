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

func (this Univalues) Clone() Univalues {
	return common.Clone(this)
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

func (this Univalues) Sort(groupIDs []uint32) Univalues {
	sortees := make([]struct {
		groupID uint32
		length  int
		str     string
		bytes   []byte
		tx      uint32
		value   interfaces.Univalue
	}, len(this))

	// t0 := time.Now()
	for i := 0; i < len(this); i++ {
		str := this[i].GetPath()
		bytes := *codec.UnsafeStringToBytes(str) // 100% faster than ([]byte(*str))

		sortees[i] = struct {
			groupID uint32
			length  int
			str     string
			bytes   []byte
			tx      uint32
			value   interfaces.Univalue
		}{
			groupID: groupIDs[i],
			length:  len(bytes),
			str:     *str,
			bytes:   bytes[ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH:],
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
		return (this[i].(*univalue.Univalue)).Less(this[j].(*univalue.Univalue))
	}

	sort.Slice(sortees, sorter)
	for i := 0; i < len(sortees); i++ {
		this[i] = sortees[i].value
		groupIDs[i] = sortees[i].groupID
	}
	return this
}

// func (this Univalues) CompressKeys(dict *ccurlcommon.Dict) {
// 	for i, univ := range this {
// 		compressedKey := (*univ.GetPath())[ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH:ccurlcommon.ETH10_ACCOUNT_FULL_LENGTH]
// 		newKey := dict.Compress(compressedKey, nil) + (*univ.GetPath())[ccurlcommon.ETH10_ACCOUNT_FULL_LENGTH:]
// 		this[i].SetPath(&newKey)
// 	}
// }

// func (this Univalues) DecompressKeys(dict *ccurlcommon.Dict) {
// 	for i := range this {
// 		key := *this[i].GetPath()
// 		idx := strings.Index(*this[i].GetPath(), "/")
// 		newKey := ccurlcommon.ETH10 + dict.Decompress(key[:idx]) + key[idx:]
// 		this[i].SetPath(&newKey)
// 	}
// }
