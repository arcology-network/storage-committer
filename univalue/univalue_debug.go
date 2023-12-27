package univalue

import (
	"crypto/sha256"
	"fmt"

	"github.com/arcology-network/common-lib/common"
	intf "github.com/arcology-network/concurrenturl/interfaces"
)

func (this *Univalue) Checksum() [32]byte {
	return sha256.Sum256(this.Encode())
}

func (this *Univalue) Print() {
	spaces := " " //fmt.Sprintf("%"+strconv.Itoa(len(strings.Split(*this.path, "/"))*1)+"v", " ")
	fmt.Print(spaces+"tx: ", this.tx)
	fmt.Print(spaces+"reads: ", this.reads)
	fmt.Print(spaces+"writes: ", this.writes)
	fmt.Print(spaces+"DeltaWrites: ", this.deltaWrites)
	fmt.Print(spaces+"persistent: ", this.persistent)
	fmt.Print(spaces+"preexists: ", this.preexists)

	fmt.Print(spaces+"path: ", *this.path, "      ")
	common.IfThenDo(this.value != nil, func() { this.value.(intf.Type).Print() }, func() { fmt.Print("nil") })
}

func (this *Univalue) Equal(other *Univalue) bool {
	if this.value == nil && other.Value() == nil {
		return true
	}

	if (this.value == nil && other.Value() != nil) || (this.value != nil && other.Value() == nil) {
		return false
	}

	vFlag := this.value.(intf.Type).Equal(other.Value().(intf.Type))
	return this.tx == other.GetTx() &&
		*this.path == *other.GetPath() &&
		this.reads == other.Reads() &&
		this.writes == other.Writes() &&
		vFlag &&
		this.preexists == other.Preexist()
}
