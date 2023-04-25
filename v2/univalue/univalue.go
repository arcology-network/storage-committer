package univalue

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	commutative "github.com/arcology-network/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/v2/type/noncommutative"
)

type Univalue struct {
	vType       uint8
	tx          uint32
	path        *string
	reads       uint32
	writes      uint32
	deltaWrites uint32
	value       interface{}
	preexists   bool
	reserved    interface{}
	reclaimFunc func(interface{})
}

func NewUnivalue(tx uint32, key string, reads, writes uint32, deltaWrites uint32, args ...interface{}) *Univalue {
	v := &Univalue{
		vType:       (&Univalue{}).GetTypeID(args[0]),
		tx:          tx,
		path:        &key,
		reads:       reads,
		writes:      writes,
		deltaWrites: deltaWrites,
		value:       args[0],
	}

	if len(args) > 1 {
		v.SetPreexist(key, args[1])
	}

	return v
}

func (this *Univalue) TypeID() uint8 { return this.vType }

func (this *Univalue) Init(tx uint32, key string, reads, writes uint32, v interface{}, args ...interface{}) {
	this.vType = (&Univalue{}).GetTypeID(v)
	this.tx = tx
	this.path = &key
	this.reads = reads
	this.writes = writes
	this.value = v
	this.preexists = false

	this.reserved = nil
	if len(args) > 0 {
		this.SetPreexist(key, args[0]) // Check if the key  exists in indexer already
		// value.composite = value.IfConcurrentWritable()
	}
}

func (this *Univalue) Reclaim() {
	if this.reclaimFunc != nil {
		this.reclaimFunc(this)
	}
}

func (*Univalue) GetTypeID(value interface{}) uint8 {
	if value != nil {
		switch value.(type) {
		case *noncommutative.Bigint:
			return ccurlcommon.NoncommutativeBigint
		case *noncommutative.Int64:
			return ccurlcommon.NoncommutativeInt64
		case *noncommutative.String:
			return ccurlcommon.NoncommutativeString
		case *noncommutative.Bytes:
			return ccurlcommon.NoncommutativeBytes
		case *commutative.Meta:
			return ccurlcommon.CommutativeMeta
		case *commutative.U256: /* Commutatives */
			return ccurlcommon.CommutativeUint256
		case *commutative.Int64:
			return ccurlcommon.CommutativeInt64
		}
	}
	return uint8(reflect.Invalid)
}

func (this *Univalue) IsCommutative() bool {
	if this.value != nil {
		switch this.value.(type) {
		case *commutative.Meta:
			return true
		case *commutative.U256: /* Commutatives */
			return true
		case *commutative.Int64:
			return true
		case *commutative.Uint64:
			return true
		}
	}
	return false
}

func (this *Univalue) GetTx() uint32                 { return this.tx }
func (this *Univalue) GetPath() *string              { return this.path }
func (this *Univalue) SetPath(path *string)          { this.path = path }
func (this *Univalue) ClearPath()                    { *this.path = (*this.path)[:0] }
func (this *Univalue) Value() interface{}            { return this.value }
func (this *Univalue) SetValue(newValue interface{}) { this.value = newValue }
func (this *Univalue) ClearReserve()                 { this.reserved = nil }

func (this *Univalue) Reads() uint32       { return this.reads }
func (this *Univalue) Writes() uint32      { return this.writes } // Exist in cache as a failed read
func (this *Univalue) DeltaWrites() uint32 { return this.deltaWrites }

func (this *Univalue) Preexist() bool { return this.preexists } // Exist in cache as a failed read

func (this *Univalue) IncrementReads(reads uint32)   { this.reads += reads }
func (this *Univalue) IncrementWrites(writes uint32) { this.writes += writes }
func (this *Univalue) IncrementDelta(writes uint32)  { this.deltaWrites += writes }

func (this *Univalue) DecrementReads() {
	if this.reads <= uint32(0) {
		panic("Reads cannot be negative !!!")
	}
	this.reads--
}

func (this *Univalue) SetPreexist(key string, source interface{}) {
	this.preexists = source.(ccurlcommon.IndexerInterface).RetriveShallow(key) != nil
}

func (this *Univalue) IfConcurrentWritable() bool { // Call this before setting the value attribute to nil
	return (this.value != nil && this.reads == 0 && this.writes == 0)
}

func (this *Univalue) Get(tx uint32, path string, source interface{}) interface{} {
	if this.value != nil {
		tempV, r, w := this.value.(ccurlcommon.TypeInterface).Get(source) //RW: Affiliated reads and writes
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

func (this *Univalue) PrecheckAttributes(other *Univalue) {
	if other.reads == 0 && other.writes == 0 && other.deltaWrites == 0 {
		panic("Error: Read/Write/Deltawrite all zero!!")
	}

	if other.writes == 0 && other.deltaWrites == 0 {
		panic("Error: Value type mismatched!") // Read only variable should never be here.
	}

	if this.preexists && this.IsCommutative() && this.Reads() > 0 && this.IfConcurrentWritable() == other.IfConcurrentWritable() {
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
		vType:     this.vType,
		tx:        this.tx,
		path:      this.path,
		reads:     this.reads,
		writes:    this.writes,
		value:     this.value,
		preexists: this.preexists,
	}

	if v.value != nil {
		v.value = this.value.(ccurlcommon.TypeInterface).Deepcopy()
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

func (this *Univalue) Equal(other *Univalue) bool {
	if this.value == nil &&
		other.value == nil &&
		this.TypeID() == other.TypeID() {
		return true
	}

	if (this.value == nil && other.value != nil) || (this.value != nil && other.value == nil) {
		return false
	}

	var vFlag bool
	if this.value.(ccurlcommon.TypeInterface).TypeID() == ccurlcommon.CommutativeMeta {
		vFlag = this.value.(*commutative.Meta).Equal(other.value.(*commutative.Meta))
	} else {
		vFlag = reflect.DeepEqual(this.value, other.value)
	}

	return this.tx == other.tx &&
		*this.path == *other.path &&
		this.reads == other.reads &&
		this.writes == other.writes &&
		vFlag &&
		this.preexists == other.preexists
}
