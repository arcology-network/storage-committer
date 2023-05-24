package univalue

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"

	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/commutative"
)

type Univalue struct {
	Unimeta
	value interface{}
	cache []byte
}

// func NewUnivalue

func NewUnivalue(tx uint32, key string, reads, writes uint32, deltaWrites uint32, args ...interface{}) *Univalue {
	return &Univalue{
		Unimeta{
			vType:       common.IfThenDo1st(args[0] != nil, func() uint8 { return args[0].(ccurlcommon.TypeInterface).TypeID() }, uint8(reflect.Invalid)),
			tx:          tx,
			path:        &key,
			reads:       reads,
			writes:      writes,
			deltaWrites: deltaWrites,
			preexists:   common.IfThenDo1st(len(args) > 1, func() bool { return (&Unimeta{}).CheckPreexist(key, args[1]) }, false),
		},
		args[0],
		[]byte{},
	}
}

func (*Univalue) New(meta, value, cache interface{}) interface{} {
	return &Univalue{
		meta.(Unimeta),
		value,
		cache.([]byte),
	}
}

func (this *Univalue) SetTx(txId uint32)             { this.tx = txId }
func (this *Univalue) ClearCache()                   { this.cache = this.cache[:0] }
func (this *Univalue) Value() interface{}            { return this.value }
func (this *Univalue) SetValue(newValue interface{}) { this.value = newValue }

func (this *Univalue) GetUnimeta() interface{} { return this.Unimeta }
func (this *Univalue) GetCache() interface{}   { return this.cache }

func (this *Univalue) Meta() ccurlcommon.UnivalueInterface {
	var v interface{}
	if this.value != nil {
		value := this.value.(ccurlcommon.TypeInterface)
		if this.deltaWrites > 0 && this.reads == 0 && this.writes == 0 && this.TypeID() != commutative.PATH {
			v = this.value.(ccurlcommon.TypeInterface).New(nil, value.Delta(), value.Sign(), value.Min(), value.Max())
		}
	}

	return &Univalue{
		this.Unimeta,
		v,
		[]byte{},
	}
}

func (this *Univalue) Init(tx uint32, key string, reads, writes uint32, v interface{}, args ...interface{}) {
	this.vType = common.IfThenDo1st(v != nil, func() uint8 { return v.(ccurlcommon.TypeInterface).TypeID() }, uint8(reflect.Invalid))
	this.tx = tx
	this.path = &key
	this.reads = reads
	this.writes = writes
	this.value = v
	this.preexists = common.IfThenDo1st(len(args) > 0, func() bool { return (&Unimeta{}).CheckPreexist(key, args[0]) }, false)
}

func (this *Univalue) Reclaim() {
	if this.reclaimFunc != nil {
		this.reclaimFunc(this)
	}
}

func (this *Univalue) Do(tx uint32, path string, do interface{}) interface{} {
	ret := do.(func(interface{}) interface{})(this).([]interface{})
	this.reads += ret[0].(uint32)
	this.writes += ret[1].(uint32)
	this.deltaWrites += ret[2].(uint32)
	return ret[3]
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

func (this *Univalue) WriteTo(writeCache ccurlcommon.WriteCacheInterface) {
	common.IfThenDo(this.writes == 0 && this.deltaWrites == 0,
		func() { writeCache.Read(this.tx, *this.GetPath()) },
		func() { writeCache.Write(this.tx, *this.GetPath(), this.value) },
	)
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
		this.value = this.value.(ccurlcommon.TypeInterface).Clone()
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
func (this *Univalue) ApplyDelta(v interface{}) error {
	vec := v.([]ccurlcommon.UnivalueInterface)

	/* Precheck & Merge attributes*/
	for i := 0; i < len(vec); i++ {
		this.PrecheckAttributes(vec[i].(*Univalue))
		this.writes += vec[i].Writes()
		this.reads += vec[i].Reads()
	}

	// Apply transitions
	var err error
	if this.Value() != nil {
		if this.value, _, err = this.Value().(ccurlcommon.TypeInterface).ApplyDelta(v); err != nil {
			return err
		}
	}
	return nil
}

func (this *Univalue) IsConcurrentWritable() bool { // Call this before setting the value attribute to nil
	return (this.value != nil && this.Writes() == 0 && this.DeltaWrites() == 0)
}

func (this *Univalue) PrecheckAttributes(other *Univalue) {
	if other.reads == 0 && other.writes == 0 && other.deltaWrites == 0 {
		panic("Error: Read/Write/Deltawrite all zero!!")
	}

	if other.writes == 0 && other.deltaWrites == 0 {
		panic("Error: Value type mismatched!") // Read only variable should never be here.
	}

	if this.preexists && this.Value().(ccurlcommon.TypeInterface).IsCommutative() && this.Reads() > 0 && this.IsConcurrentWritable() == other.IsConcurrentWritable() {
		this.Print()
		fmt.Println("================================================================")
		other.Print()
		panic("Error: The composite attribute must match in different transitions")
	}

	if this.Value() == nil && this.IsConcurrentWritable() {
		panic("Error: A deleted value cann't be composite")
	}

	if !this.preexists && this.IsConcurrentWritable() {
		panic("Error: A new value cann't be composite")
	}
}

func (this *Univalue) Clone() interface{} {
	v := &Univalue{
		this.Unimeta.Clone(),
		common.IfThenDo1st(this.value != nil, func() interface{} { return this.value.(ccurlcommon.TypeInterface).Clone() }, this.value),
		common.Clone(this.cache),
	}
	return v
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
	fmt.Print(spaces+"path: ", *this.path)
	// fmt.Print(spaces+"value: ", this.value)
	fmt.Println(spaces+"preexists: ", this.preexists)

	//this.value.(ccurlcommon.TypeInterface).Print()
	// fmt.Println("--------------------------------------------------------")
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
