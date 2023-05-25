package univalue

import (
	"bytes"
	"crypto/sha256"
	"sort"

	codec "github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
)

type Univalues []ccurlcommon.UnivalueInterface

func (this Univalues) To(filters ...func(ccurlcommon.UnivalueInterface) ccurlcommon.UnivalueInterface) Univalues {
	for _, condition := range filters {
		this = common.CastTo(this, condition)
	}
	common.RemoveIf((*[]ccurlcommon.UnivalueInterface)(&this), func(v ccurlcommon.UnivalueInterface) bool { return v == nil })
	return this
}

// Debugging only
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

func (this Univalues) SortByPath() Univalues {
	sort.SliceStable(this, func(i, j int) bool {
		if len(*this[i].GetPath()) != len(*this[j].GetPath()) {
			return len(*this[i].GetPath()) < len(*this[j].GetPath())
		}
		return bytes.Equal([]byte(*this[i].GetPath()), []byte(*this[j].GetPath()))
	})
	return this
}

func (this Univalues) SortByCost() Univalues {
	sort.SliceStable(this, func(i, j int) bool {
		if (this[i].Value() == nil || this[j].Value() == nil) && (this[i].Value() != this[j].Value()) {
			return this[i].Value() == nil
		}

		if this[i].GetTx() != this[j].GetTx() {
			return this[i].GetTx() < this[j].GetTx()
		}

		if this[i].Writes() != this[j].Writes() {
			return this[i].Writes() > this[j].Writes()
		}

		if this[i].Reads() != this[j].Reads() {
			return this[i].Reads() > this[j].Reads()
		}

		if this[i].DeltaWrites() != this[j].DeltaWrites() {
			return this[i].DeltaWrites() > this[j].DeltaWrites()
		}

		if this[i].Preexist() != this[j].Preexist() {
			return this[j].Preexist()
		}

		if (!this[i].Preexist() || !this[j].Preexist()) && (this[i].Preexist() != this[j].Preexist()) {
			return this[i].Preexist()
		}

		return true
	})
	return this
}

func (this Univalues) Sort() Univalues {
	summed := make([]uint64, len(this))
	for i := ccurlcommon.ETH10_ACCOUNT_LENGTH; i < len(this); i++ {
		summed[i] = codec.String(*this[i].GetPath()).Sum()
	}

	sort.Slice(this, func(i, j int) bool {
		if summed[i] != summed[j] {
			return summed[i] < summed[j]
		}

		if *this[i].GetPath() != *this[j].GetPath() {
			return bytes.Compare([]byte(*this[i].GetPath()), []byte(*this[j].GetPath())) < 0
		}

		if (this[i].Value() == nil || this[j].Value() == nil) && (this[i].Value() != this[j].Value()) {
			return this[i].Value() == nil
		}

		if this[i].Writes() != this[j].Writes() {
			return this[i].Writes() > this[j].Writes()
		}

		if this[i].Reads() != this[j].Reads() {
			return this[i].Reads() > this[j].Reads()
		}

		if this[i].DeltaWrites() != this[j].DeltaWrites() {
			return this[i].DeltaWrites() > this[j].DeltaWrites()
		}

		if (!this[i].Preexist() || !this[j].Preexist()) && (this[i].Preexist() != this[j].Preexist()) {
			return this[i].Preexist()
		}

		if this[i].GetTx() != this[j].GetTx() {
			return this[i].GetTx() < this[j].GetTx()
		}
		return true
	})
	return this
}
