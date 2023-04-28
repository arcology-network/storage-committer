package univalue

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
)

type Univalue struct {
	Unimeta
	value interface{}
	cache []byte
}

func NewUnivalue(tx uint32, key string, reads, writes uint32, deltaWrites uint32, args ...interface{}) *Univalue {
	v := Univalue{
		Unimeta{
			vType:       (&Univalue{}).GetTypeID(args[0]),
			tx:          tx,
			path:        &key,
			reads:       reads,
			writes:      writes,
			deltaWrites: deltaWrites,
			preexists:   common.IfThenDo1st(args != nil && len(args) > 1, func() bool { return (&Unimeta{}).CheckPreexist(key, args[1]) }, false),
		},
		args[0],
		[]byte{},
	}
	return &v
}

func (this *Univalue) ClearCache()                   { this.cache = this.cache[:0] }
func (this *Univalue) Value() interface{}            { return this.value }
func (this *Univalue) SetValue(newValue interface{}) { this.value = newValue }

func (this *Univalue) Init(tx uint32, key string, reads, writes uint32, v interface{}, args ...interface{}) {
	this.vType = (&Univalue{}).GetTypeID(v)
	this.tx = tx
	this.path = &key
	this.reads = reads
	this.writes = writes
	this.value = v
	this.preexists = common.IfThenDo1st(args != nil && len(args) > 0, func() bool { return (&Unimeta{}).CheckPreexist(key, args[0]) }, false)
}

func (this *Univalue) Reclaim() {
	if this.reclaimFunc != nil {
		this.reclaimFunc(this)
	}
}

func (this *Univalue) Get(tx uint32, path string, source interface{}) interface{} {
	if this.value != nil {
		tempV, r, w := this.value.(ccurlcommon.TypeInterface).Get() //RW: Affiliated reads and writes
		this.reads += r
		this.writes += w
		return tempV
	}
	this.IncrementReads(1)
	return this.value
}

func (this *Univalue) Set(tx uint32, path string, typedV interface{}, indexer interface{}) error { // update the value
	this.tx = tx
	if this.Value() == nil && typedV == nil {
		this.writes++ // Delete an non-existing value
		return errors.New("Error: The value doesn't exists")
	}

	if this.Value() == nil { // Added a new value or try to delete an non-existent value
		this.vType = typedV.(ccurlcommon.TypeInterface).TypeID()
		v, r, w, dw := typedV.(ccurlcommon.TypeInterface).CopyTo(typedV)
		this.value = v
		this.writes += w
		this.reads += r
		this.deltaWrites += dw
		return nil
	}

	if this.writes == 0 && this.value != nil && typedV != nil { // Make a deep copy if haven't done so
		this.value = this.value.(ccurlcommon.TypeInterface).Deepcopy()
	}

	v, r, w, dw, err := this.value.(ccurlcommon.TypeInterface).Set(typedV, []interface{}{path, *this.path, tx, indexer}) // Update one the current value
	this.value = v
	this.writes += w
	this.reads += r
	this.deltaWrites += dw

	if typedV == nil && this.Value().(ccurlcommon.TypeInterface).IsSelf(path) { // Delete the entry but keep the access record.
		this.vType = uint8(reflect.Invalid)
		this.value = typedV // Delete the value
		this.writes++
	}
	return err
}

// Check & Merge attributes
func (this *Univalue) ApplyDelta(tx uint32, v interface{}) error {
	vec := v.([]ccurlcommon.UnivalueInterface)

	/* Precheck & Merge attributes*/
	for i := 0; i < len(vec); i++ {
		this.PrecheckAttributes(vec[i].(*Univalue))
		this.writes += vec[i].Writes()
		this.reads += vec[i].Reads()
	}

	// Apply transitions
	if this.Value() != nil {
		this.value = this.Value().(ccurlcommon.TypeInterface).ApplyDelta(v)
	}
	return nil
}

func (this *Univalue) IfConcurrentWritable() bool { // Call this before setting the value attribute to nil
	return (this.value != nil && this.reads == 0 && this.writes == 0)
}

func (this *Univalue) PrecheckAttributes(other *Univalue) {
	if other.reads == 0 && other.writes == 0 && other.deltaWrites == 0 {
		panic("Error: Read/Write/Deltawrite all zero!!")
	}

	if other.writes == 0 && other.deltaWrites == 0 {
		panic("Error: Value type mismatched!") // Read only variable should never be here.
	}

	if this.preexists && this.IsCommutative(this) && this.Reads() > 0 && this.IfConcurrentWritable() == other.IfConcurrentWritable() {
		this.Print()
		fmt.Println("================================================================")
		other.Print()
		panic("Error: The composite attribute must match in different transitions")
	}

	if this.Value() == nil && this.IfConcurrentWritable() {
		panic("Error: A deleted value cann't be composite")
	}

	if !this.preexists && this.IfConcurrentWritable() {
		panic("Error: A new value cann't be composite")
	}
}

func (this *Univalue) Deepcopy() interface{} {
	v := &Univalue{
		this.Unimeta.Clone(),
		common.IfThen(this.value != nil, this.value.(ccurlcommon.TypeInterface).Deepcopy(), nil),
		codec.Bytes(this.cache).Clone(),
	}
	return v
}

func (this *Univalue) Checksum() [32]byte {
	return sha256.Sum256(this.Encode())
}

func (this *Univalue) Print() {
	spaces := fmt.Sprintf("%"+strconv.Itoa(len(strings.Split(*this.path, "/"))*4)+"v", " ")
	fmt.Println(spaces+"tx: ", this.tx)
	fmt.Println(spaces+"reads: ", this.reads)
	fmt.Println(spaces+"writes: ", this.writes)
	fmt.Println(spaces+"path: ", *this.path)
	fmt.Println(spaces+"value: ", this.value)
	fmt.Println(spaces+"preexists: ", this.preexists)

	//this.value.(ccurlcommon.TypeInterface).Print()
	fmt.Println("--------------------------------------------------------")
}

func (this *Univalue) Equal(other ccurlcommon.UnivalueInterface) bool {
	if this.value == nil && other.Value() == nil {
		return true
	}

	if (this.value == nil && other.Value() != nil) || (this.value != nil && other.Value() == nil) {
		return false
	}

	vFlag := this.value.(ccurlcommon.TypeInterface).Equal(other.Value().(ccurlcommon.TypeInterface))
	return this.tx == other.GetTx() &&
		*this.path == *other.GetPath() &&
		this.reads == other.Reads() &&
		this.writes == other.Writes() &&
		vFlag &&
		this.preexists == other.Preexist()
}
