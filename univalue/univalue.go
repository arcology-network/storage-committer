package univalue

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type Univalue struct {
	Unimeta
	value interface{}
	cache []byte
}

// func NewUnivalue

func NewUnivalue(tx uint32, key string, reads, writes uint32, deltaWrites uint32, v interface{}, source interface{}) *Univalue {
	return &Univalue{
		Unimeta{
			vType:       common.IfThenDo1st(v != nil, func() uint8 { return v.(interfaces.Type).TypeID() }, uint8(reflect.Invalid)),
			tx:          tx,
			path:        &key,
			reads:       reads,
			writes:      writes,
			deltaWrites: deltaWrites,
			preexists:   common.IfThenDo1st(source != nil, func() bool { return (&Unimeta{}).CheckPreexist(key, source) }, false),
		},
		v,
		[]byte{},
	}
}

func (*Univalue) New(meta, value, cache interface{}) interface{} {
	return &Univalue{
		*meta.(*Unimeta),
		value,
		cache.([]byte),
	}
}

func (this *Univalue) From(v interfaces.Univalue) interface{} { return v }

// func (this *Univalue) IsHotLoaded() bool             { return this.reads > 1 }
func (this *Univalue) SetTx(txId uint32)  { this.tx = txId }
func (this *Univalue) ClearCache()        { this.cache = this.cache[:0] }
func (this *Univalue) Value() interface{} { return this.value }
func (this *Univalue) SetValue(newValue interface{}) interfaces.Univalue {
	if this.value != nil && reflect.TypeOf(this.value) != reflect.TypeOf(newValue) && newValue != nil {
		panic("Wrong type")
	}
	this.value = newValue
	return this
}

func (this *Univalue) GetUnimeta() interface{} { return &this.Unimeta }
func (this *Univalue) GetCache() interface{}   { return this.cache }

func (this *Univalue) Init(tx uint32, key string, reads, writes, deltaWrites uint32, v interface{}, args ...interface{}) *Univalue {
	this.vType = common.IfThenDo1st(v != nil, func() uint8 { return v.(interfaces.Type).TypeID() }, uint8(reflect.Invalid))
	this.tx = tx
	this.path = &key
	this.reads = reads
	this.writes = writes
	this.deltaWrites = deltaWrites
	this.value = v
	this.preexists = common.IfThenDo1st(len(args) > 0, func() bool { return (&Unimeta{}).CheckPreexist(key, args[0]) }, false)
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
		tempV, r, w := this.value.(interfaces.Type).Get() //RW: Affiliated reads and writes
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
		Find(string, interface{}) (interface{}, interface{})
	})

	common.IfThenDo(this.writes == 0 && this.deltaWrites == 0,
		func() { writeCache.Read(this.tx, *this.GetPath(), this.value) }, // Add reads
		func() { writeCache.Write(this.tx, *this.GetPath(), this.value) },
	)

	_, univ := writeCache.Find(*this.GetPath(), nil)
	readsDiff := this.Reads() - univ.(interfaces.Univalue).Reads()
	writesDiff := this.Writes() - univ.(interfaces.Univalue).Writes()
	deltaWriteDiff := this.DeltaWrites() - univ.(interfaces.Univalue).DeltaWrites()

	univ.(interfaces.Univalue).IncrementReads(readsDiff)
	univ.(interfaces.Univalue).IncrementWrites(writesDiff)
	univ.(interfaces.Univalue).IncrementDeltaWrites(deltaWriteDiff)
}

func (this *Univalue) Set(tx uint32, path string, typedV interface{}, importer interface{}) error { // update the value
	this.tx = tx
	if this.Value() == nil && typedV == nil {
		this.writes++ // Delete an non-existing value
		return errors.New("Error: The value doesn't exists")
	}

	if this.Value() == nil { // Added a new value or try to delete an non-existent value
		this.vType = typedV.(interfaces.Type).TypeID()
		v, r, w, dw := typedV.(interfaces.Type).CopyTo(typedV)
		this.value = v
		this.writes += w
		this.reads += r
		this.deltaWrites += dw
		return nil
	}

	if this.writes == 0 && this.value != nil && typedV != nil { // Make a deep copy if haven't done so
		this.value = this.value.(interfaces.Type).Clone()
	}

	v, r, w, dw, err := this.value.(interfaces.Type).Set(typedV, []interface{}{path, *this.path, tx, importer}) // Update one the current value
	this.value = v
	this.writes += w
	this.reads += r
	this.deltaWrites += dw

	if typedV == nil && this.Value().(interfaces.Type).IsSelf(path) { // Delete the entry but keep the access record.
		this.vType = uint8(reflect.Invalid)
		this.value = typedV // Delete the value
		this.writes++
	}
	return err
}

// Check & Merge attributes
func (this *Univalue) ApplyDelta(v interface{}) error {
	vec := v.([]interfaces.Univalue)

	/* Precheck & Merge attributes*/
	for i := 0; i < len(vec); i++ {
		this.PrecheckAttributes(vec[i].(*Univalue))
		this.writes += vec[i].Writes()
		this.reads += vec[i].Reads()
		this.deltaWrites += vec[i].DeltaWrites()
	}

	// Apply transitions
	var err error
	if this.Value() != nil {
		if this.value, _, err = this.Value().(interfaces.Type).ApplyDelta(v); err != nil {
			return err
		}
	}
	return nil
}

func (this *Univalue) IsConcurrentWritable() bool { // Call this before setting the value attribute to nil
	return (this.Writes() == 0 && this.Reads() == 0)
}

func (this *Univalue) PrecheckAttributes(other *Univalue) {
	if other.reads == 0 && other.writes == 0 && other.deltaWrites == 0 {
		panic("Error: Read/Write/Deltawrite all zero!!")
	}

	if other.writes == 0 && other.deltaWrites == 0 {
		panic("Error: Value type mismatched!") // Read only variable should never be here.
	}

	if this.preexists && this.Value().(interfaces.Type).IsCommutative() && this.Reads() > 0 && this.IsConcurrentWritable() == other.IsConcurrentWritable() {
		this.Print()
		fmt.Println("================================================================")
		other.Print()
		panic("Error: The composite attribute must match in different transitions")
	}

	if this.Value() == nil && this.IsConcurrentWritable() {
		// panic("Error: A deleted value cann't be composite")
	}

	if !this.preexists && this.IsConcurrentWritable() {
		panic("Error: A new value cann't be composite")
	}
}

func (this *Univalue) Clone() interface{} {
	v := &Univalue{
		this.Unimeta.Clone(),
		common.IfThenDo1st(this.value != nil, func() interface{} { return this.value.(interfaces.Type).Clone() }, this.value),
		common.Clone(this.cache),
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
