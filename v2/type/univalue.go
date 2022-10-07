package ccurltype

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
	commutative "github.com/HPISTechnologies/concurrenturl/v2/type/commutative"
	noncommutative "github.com/HPISTechnologies/concurrenturl/v2/type/noncommutative"
)

type Univalue struct {
	transitType uint8 // Transition type
	vType       uint8
	tx          uint32
	path        *string
	reads       uint32
	writes      uint32
	value       interface{}
	preexists   bool
	composite   bool
	reserved    interface{}
	reclaimFunc func(interface{})

	inited bool
}

func NewUnivalue(transitType uint8, tx uint32, key string, reads, writes uint32, args ...interface{}) *Univalue {
	v := &Univalue{
		transitType: transitType,
		vType:       (&Univalue{}).GetTypeID(args[0]),
		tx:          tx,
		path:        &key,
		reads:       reads,
		writes:      writes,
		value:       args[0],
		preexists:   false,
		composite:   false,
	}

	if len(args) > 1 {
		v.SetPreexist(key, args[1])
		v.composite = v.IfComposite()
	}

	return v
}

func (value *Univalue) Init(transitType uint8, tx uint32, key string, reads, writes uint32, v interface{}, args ...interface{}) {
	if value.inited {
		panic(fmt.Sprintf("Univalue.Init called twice.\n\toriginal obj = %v\n\tvalue = %v\n\treserved = %v", value, value.value, value.reserved))
	}

	value.transitType = transitType
	value.vType = (&Univalue{}).GetTypeID(v)
	value.tx = tx
	value.path = &key
	value.reads = reads
	value.writes = writes
	value.value = v
	value.preexists = false
	value.composite = false
	value.reserved = nil
	if len(args) > 0 {
		value.SetPreexist(key, args[0]) // Check if the key  exists in indexer already
		value.composite = value.IfComposite()
	}
	value.inited = true
}

func (value *Univalue) Reclaim() {
	if value.reclaimFunc != nil {
		value.reclaimFunc(value)
	}
	value.inited = false
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
		case *commutative.Balance: /* Commutatives */
			return ccurlcommon.CommutativeBalance
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
		case *commutative.Balance: /* Commutatives */
			return true
		case *commutative.Int64:
			return true
		}
	}
	return false
}

func (this *Univalue) GetTx() uint32                 { return this.tx }
func (this *Univalue) GetPath() *string              { return this.path }
func (this *Univalue) SetPath(path string)           { this.path = &path }
func (this *Univalue) ClearPath()                    { *this.path = (*this.path)[:0] }
func (this *Univalue) Value() interface{}            { return this.value }
func (this *Univalue) SetValue(newValue interface{}) { this.value = newValue }
func (this *Univalue) ClearReserve()                 { this.reserved = nil }

func (this *Univalue) GetTransitionType() uint8       { return this.transitType }
func (this *Univalue) SetTransitionType(typeID uint8) { this.transitType = typeID }

func (this *Univalue) Reads() uint32   { return this.reads }
func (this *Univalue) Writes() uint32  { return this.writes }    // Exist in cache as a failed read
func (this *Univalue) Preexist() bool  { return this.preexists } // Exist in cache as a failed read
func (this *Univalue) Composite() bool { return this.composite }

func (this *Univalue) IncrementRead()  { this.reads++ }
func (this *Univalue) IncrementWrite() { this.writes++ }

func (this *Univalue) SetPreexist(key string, source interface{}) {
	this.preexists = source.(ccurlcommon.IndexerInterface).RetriveShallow(key) != nil
}

func (this *Univalue) IfComposite() bool { // Call this before setting the value attribute to nil
	if this.value != nil && this.Preexist() {
		return this.value.(ccurlcommon.TypeInterface).Composite() && this.reads == 0
	}
	return false
}

// Update the parent meta if necessary
func (this *Univalue) UpdateParentMeta(tx uint32, value interface{}, source interface{}) bool {
	if this.Writes() == 0 {
		this.value = this.Value().(ccurlcommon.TypeInterface).Deepcopy()
	}

	child := value.(*Univalue)
	meta := this.Value().(*commutative.Meta)
	if meta.UpdateCaches(tx, child, source) {
		this.IncrementWrite()
		return true
	}
	return false
}

func (this *Univalue) Export(source interface{}) (interface{}, interface{}) {
	if this.Value() != nil {
		this.value = this.value.(ccurlcommon.TypeInterface).Delta(source.(ccurlcommon.IndexerInterface).Buffer())
	}

	accessRecord := &Univalue{ // For the arbitrator, just make a deep copy and clear the value field
		tx:        this.GetTx(),
		vType:     this.GetTypeID(this.Value()),
		path:      this.GetPath(),
		reads:     this.Reads(),
		writes:    this.Writes(),
		value:     this.Value(),
		preexists: this.Preexist(),
		composite: this.IfComposite(),
	}

	if accessRecord.Value() != nil {
		accessRecord.value = accessRecord.Value().(ccurlcommon.TypeInterface).ToAccess()
	}

	if this.Writes() > 0 && this.Value() != nil { // Rewrite an existing entry or create a new one
		return accessRecord, this
	}

	if this.Writes() > 0 && this.Value() == nil && this.Preexist() { // Deletion of an existing entry
		return accessRecord, this
	}

	return accessRecord, nil
}

func (this *Univalue) Get(tx uint32, path string, source interface{}) interface{} {
	if this.value != nil {
		tempV, r, w := this.value.(ccurlcommon.TypeInterface).Get(tx, path, source) //RW: Affiliated reads and writes
		this.reads += r
		this.writes += w
		this.composite = tempV.(ccurlcommon.TypeInterface).Composite()
		return tempV
	}
	this.IncrementRead()
	return this.value
}

func (this *Univalue) Peek(source interface{}) interface{} {
	if this.value != nil {
		return this.value.(ccurlcommon.TypeInterface).Peek(source)
	}
	return this.value
}

func (this *Univalue) Set(tx uint32, path string, value interface{}, source interface{}) error { // update the value
	this.tx = tx
	this.writes++
	if this.Value() != nil && value != nil && this.vType != value.(ccurlcommon.TypeInterface).TypeID() {
		return errors.New("Error: Types don't match !")
	}

	if this.Value() == nil && value != nil { // A New value
		this.vType = value.(ccurlcommon.TypeInterface).TypeID()
		this.value = value
		return nil
	}
	this.writes-- // Reset writes

	if this.writes == 0 && this.value != nil && value != nil { // Make a deep copy if haven't yet
		this.value = this.value.(ccurlcommon.TypeInterface).Deepcopy()
	}

	r, w, err := this.value.(ccurlcommon.TypeInterface).Set(tx, path, value, source) // Update the current value
	this.writes += w
	this.reads += r

	if value == nil { // Delete an entry
		this.vType = uint8(reflect.Invalid)
		this.value = value
	}
	return err
}

// Check & Merge attributes
func (this *Univalue) ApplyDelta(tx uint32, v interface{}) error {
	vec := v.([]ccurlcommon.UnivalueInterface)

	/* Precheck & Merge attributes*/
	for i := 0; i < len(vec); i++ {
		this.PrecheckAttributes(this, vec[i].(*Univalue))
		this.writes += vec[i].Writes()
		this.reads += vec[i].Reads()
		this.composite = this.composite && vec[i].Composite()
	}

	// Apply transitions
	if this.Value() != nil {
		this.value = this.Value().(ccurlcommon.TypeInterface).ApplyDelta(tx, v)
	}
	return nil
}

func (*Univalue) PrecheckAttributes(this *Univalue, other *Univalue) {
	if other.writes == 0 {
		panic("Error: Value type mismatched!")
	}

	if this.composite != other.composite &&
		this.Value().(ccurlcommon.TypeInterface).TypeID() != ccurlcommon.CommutativeMeta {
		this.Print()
		fmt.Println("================================================================")
		other.Print()
		panic("Error: The composite attribute must match in different transitions")
	}

	if this.Value() == nil && this.composite {
		panic("Error: A deleted value cann't be composite")
	}

	if !this.preexists && this.composite {
		panic("Error: A new value cann't be composite")
	}
}

func (this *Univalue) PostcheckAttributes() error {
	if this.composite && this.reads > 0 {
		panic("Error: Inconsistent properites")
	}
	return nil
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
		composite: this.composite,
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
	fmt.Println(spaces+"composite: ", this.composite)
	//this.value.(ccurlcommon.TypeInterface).Print()
	fmt.Println("--------------------------------------------------------")
}

func (this *Univalue) Equal(other *Univalue) bool {
	if this.transitType == other.transitType &&
		this.vType == other.vType &&
		this.tx == other.tx &&
		this.path == other.path &&
		this.reads == other.reads &&
		this.writes == other.writes &&
		this.value == other.value &&
		this.preexists == other.preexists {
		return true
	}
	return false
}

func (this *Univalue) EqualTransition(other *Univalue) bool {
	var vFlag bool
	if this.value != nil && this.value.(ccurlcommon.TypeInterface).TypeID() == ccurlcommon.CommutativeMeta {
		vFlag = this.value.(*commutative.Meta).Equal(other.value.(*commutative.Meta))
	} else {
		vFlag = reflect.DeepEqual(this.value, other.value)
	}

	return this.transitType == other.transitType &&
		this.tx == other.tx &&
		*this.path == *other.path &&
		this.reads == other.reads &&
		this.writes == other.writes &&
		vFlag &&
		this.preexists == other.preexists
}
