package univalue

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	intf "github.com/arcology-network/storage-committer/interfaces"
)

// THe univalue is a combination of a value and a property field that contains the access information about the value.
type Univalue struct {
	Property
	value interface{}
	cache []byte
}

func NewUnivalue(tx uint32, key string, reads, writes uint32, deltaWrites uint32, v interface{}, source interface{}) *Univalue {
	return &Univalue{
		Property{
			vType:       common.IfThenDo1st(v != nil, func() uint8 { return v.(intf.Type).TypeID() }, uint8(reflect.Invalid)),
			tx:          tx,
			path:        &key,
			reads:       reads,
			writes:      writes,
			deltaWrites: deltaWrites,
			preexists:   common.IfThenDo1st(source != nil, func() bool { return (&Property{}).CheckPreexist(key, source) }, false),
		},
		v,
		[]byte{},
	}
}

func (*Univalue) New(meta, value, cache interface{}) *Univalue {
	return &Univalue{
		*meta.(*Property),
		value,
		cache.([]byte),
	}
}

func (*Univalue) Reset(this *Univalue) {
	this.Property.Reset()
	this.ClearCache()
	this.value = nil
}

func (this *Univalue) From(v *Univalue) interface{} { return v }

// func (this *Univalue) IsHotLoaded() bool             { return this.reads > 1 }
func (this *Univalue) SetTx(txId uint32)  { this.tx = txId }
func (this *Univalue) ClearCache()        { this.cache = this.cache[:0] }
func (this *Univalue) Value() interface{} { return this.value }
func (this *Univalue) SetValue(newValue interface{}) *Univalue {
	if this.value != nil && reflect.TypeOf(this.value) != reflect.TypeOf(newValue) && newValue != nil {
		panic("Wrong type")
	}

	this.value = newValue
	return this
}

func (this *Univalue) GetCache() interface{} { return this.cache }

func (this *Univalue) Init(tx uint32, key string, reads, writes, deltaWrites uint32, v interface{}, args ...interface{}) *Univalue {
	this.vType = common.IfThenDo1st(v != nil, func() uint8 { return v.(intf.Type).TypeID() }, uint8(reflect.Invalid))
	this.tx = tx
	this.path = &key
	this.reads = reads
	this.writes = writes
	this.deltaWrites = deltaWrites
	this.value = v
	this.preexists = common.IfThenDo1st(len(args) > 0, func() bool { return (&Property{}).CheckPreexist(key, args[0]) }, false)
	return this
}

func (this *Univalue) Reclaim() {
	if this.reclaimFunc != nil {
		this.reclaimFunc(this)
	}
}

func (this *Univalue) Do(tx uint32, path string, doer interface{}) interface{} {
	r, w, dw, ret := doer.(func(interface{}) (uint32, uint32, uint32, interface{}))(this)
	this.reads += r
	this.writes += w
	this.deltaWrites += dw
	return ret
}

func (this *Univalue) Get(tx uint32, path string, source interface{}) interface{} {
	if this.value != nil {
		tempV, r, w := this.value.(intf.Type).Get() //RW: Affiliated reads and writes
		this.reads += r
		this.writes += w
		return tempV
	}
	this.IncrementReads(1)
	return this.value
}

func (this *Univalue) CopyTo(writable interface{}) {
	writeCache := writable.(interface {
		Read(uint32, string, interface{}) (interface{}, interface{}, uint64)
		Write(uint32, string, interface{}) (int64, error)
		Find(uint32, string, interface{}) (interface{}, interface{})
	})

	if this.writes == 0 && this.deltaWrites == 0 {
		writeCache.Read(this.tx, *this.GetPath(), this.value)
	} else {
		writeCache.Write(this.tx, *this.GetPath(), this.value)
	}

	// here !!!
	_, univ := writeCache.Find(this.tx, *this.GetPath(), nil)
	univ.(*Univalue).IncrementReads(this.Reads())
	univ.(*Univalue).IncrementWrites(this.Writes())
	univ.(*Univalue).IncrementDeltaWrites(this.DeltaWrites())
}

func (this *Univalue) Set(tx uint32, path string, newV interface{}, inCache bool, importer interface{}) error { // update the value
	this.tx = tx
	if this.value == nil && newV == nil {
		this.writes++ // Delete an non-existing value
		return errors.New("Error: The value doesn't exists")
	}

	if this.value == nil { // Added a new value or try to delete an non-existent value
		this.vType = newV.(intf.Type).TypeID()
		v, r, w, dw := newV.(intf.Type).CopyTo(newV)
		this.value = v
		this.writes += w
		this.reads += r
		this.deltaWrites += dw
		return nil
	}

	// Write != 0 means the value has been not modified, so we don't need to make a deep copy.
	// this.value == nil, this is a new value assignment, so we don't need to make a deep copy.
	// typedV == nil, this is a delete operation, so we don't need to make a deep copy.
	// In cascading write cache, the values' access info will stripped off, so it wouldn't introduce interference.
	if this.writes == 0 && this.value != nil && newV != nil { // Make a deep copy if has't done so
		this.value = this.value.(intf.Type).Clone()
	}

	oldV := this.value.(intf.Type)
	v, r, w, dw, err := oldV.Set(newV, []interface{}{path, *this.path, tx, importer}) // Update the current value
	this.value = v
	this.writes += w
	this.reads += r
	this.deltaWrites += dw

	if newV == nil && this.Value().(intf.Type).IsSelf(path) { // Delete the entry but keep the access record.
		this.vType = uint8(reflect.Invalid)
		this.value = newV // Delete the value
		this.writes++
	}
	return err
}

// Check & Merge attributes
func (this *Univalue) ApplyDelta(vec []*Univalue) error {
	// vec := v.([]*Univalue)

	/* Precheck & Merge attributes*/
	for i := 0; i < len(vec); i++ {
		this.PrecheckAttributes(vec[i])
		this.writes += vec[i].Writes()
		this.reads += vec[i].Reads()
		this.deltaWrites += vec[i].DeltaWrites()
	}

	// Apply transitions
	typedVals := slice.Transform(vec, func(_ int, v *Univalue) intf.Type {
		if v.Value() != nil {
			return v.Value().(intf.Type)
		}
		return nil
	})

	var err error
	if this.Value() != nil {
		if this.value, _, err = this.Value().(intf.Type).ApplyDelta(typedVals); err != nil {
			return err
		}
	}
	return nil
}

func (this *Univalue) IsReadOnly() bool       { return (this.writes == 0 && this.deltaWrites == 0) }
func (this *Univalue) IsDeltaWriteOnly() bool { return (this.reads == 0 && this.writes == 0) }

func (this *Univalue) PrecheckAttributes(other *Univalue) {
	if other.reads == 0 && other.writes == 0 && other.deltaWrites == 0 {
		panic("Error: Read/Write/Deltawrite all zero!!")
	}

	if other.writes == 0 && other.deltaWrites == 0 {
		panic("Error: Value type mismatched!") // Read only variable should never be here.
	}

	if this.GetTx() != other.GetTx() &&
		this.preexists &&
		this.Value().(intf.Type).IsCommutative() &&
		this.Reads() > 0 &&
		this.IsDeltaWriteOnly() == other.IsDeltaWriteOnly() {
		this.Print()
		fmt.Println("================================================================")
		other.Print()
		panic("Error: The composite attribute must match in different transitions")
	}

	if this.Value() == nil && this.IsDeltaWriteOnly() {
		panic("Error: A deleted value cann't be composite")
	}

	if !this.preexists && this.IsDeltaWriteOnly() {
		panic("Error: A new value cann't be composite")
	}
}

func (this *Univalue) Clone() interface{} {
	v := &Univalue{
		this.Property.Clone(),
		common.IfThenDo1st(this.value != nil, func() interface{} { return this.value.(intf.Type).Clone() }, this.value),
		slice.Clone(this.cache),
	}
	return v
}

func (this *Univalue) Less(other *Univalue) bool {
	if (this.value == nil || other.value == nil) && (this.value != other.value) {
		return this.value == nil
	}

	if this.writes != other.writes {
		return this.writes > other.writes
	}

	if this.reads != other.reads {
		return this.reads > other.reads
	}

	if this.deltaWrites != other.deltaWrites {
		return this.deltaWrites > other.deltaWrites
	}

	if (!this.preexists || !other.preexists) && (this.preexists != other.preexists) {
		return this.preexists
	}
	return true
}

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

	path := *this.path
	if index := strings.Index(path, "container/"); index != -1 {
		path = path[:index] + "container/" + hex.EncodeToString([]byte(path[index:]))
	}

	fmt.Print(spaces+"path: ", path, "      ")
	common.IfThenDo(this.value != nil, func() { this.value.(intf.Type).Print() }, func() { fmt.Print("nil") })
	fmt.Println()
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
