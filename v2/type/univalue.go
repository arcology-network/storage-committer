package urltype

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	codec "github.com/arcology/common-lib/codec"
	encoding "github.com/arcology/common-lib/encoding"
	ccurlcommon "github.com/arcology/concurrenturl/v2/common"
	commutative "github.com/arcology/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology/concurrenturl/v2/type/noncommutative"
)

type Univalue struct {
	vType       uint8
	tx          uint32
	path        string
	reads       uint32
	writes      uint32
	value       interface{}
	preexists   bool
	addOrDelete bool
	composite   bool
}

func CreateUnivalueForTest(tx uint32, path string, reads, writes uint32, value interface{}, preexists, composite bool) *Univalue {
	return &Univalue{
		tx:        tx,
		path:      path,
		reads:     reads,
		writes:    writes,
		value:     value,
		preexists: preexists,
		composite: composite,
	}
}

func NewUnivalue(tx uint32, key string, reads, writes uint32, args ...interface{}) *Univalue {
	v := &Univalue{
		vType:       (&Univalue{}).GetTypeID(args[0]),
		tx:          tx,
		path:        key,
		reads:       reads,
		writes:      writes,
		value:       args[0],
		preexists:   false,
		addOrDelete: false,
		composite:   false,
	}

	if len(args) > 1 {
		v.SetPreexist(args[1])
		v.addOrDelete = v.IfAddOrDelete()
		v.composite = v.IfComposite()
	}

	return v
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

func (this *Univalue) GetTx() uint32      { return this.tx }
func (this *Univalue) GetPath() string    { return this.path }
func (this *Univalue) Value() interface{} { return this.value }
func (this *Univalue) Reads() uint32      { return this.reads }
func (this *Univalue) Writes() uint32     { return this.writes }    // Exist in cache as a failed read
func (this *Univalue) Preexist() bool     { return this.preexists } // Exist in cache as a failed read
func (this *Univalue) Composite() bool    { return this.composite }

func (this *Univalue) IncrementRead()  { this.reads++ }
func (this *Univalue) IncrementWrite() { this.writes++ }

func (this *Univalue) SetPreexist(source interface{}) {
	// this.preexists = source.(ccurlcommon.LocalCacheInterface).IfExists(this.path)
	this.preexists = source.(ccurlcommon.LocalCacheInterface).RetriveShallow(this.path) != nil
}

func (this *Univalue) IfAddOrDelete() bool {
	return (!this.Preexist() && this.Writes() > 0) || (this.Preexist() && this.Value() == nil)
}

func (this *Univalue) IfComposite() bool { // Call this before setting the value attribute to nil
	if this.value != nil && this.Preexist() {
		return this.value.(ccurlcommon.TypeInterface).Composite() && this.reads == 0
	}
	return false
}

// Update the parent meta if necessary
func (this *Univalue) UpdateParentMeta(tx uint32, value interface{}, source interface{}) error {
	if this.Value().(ccurlcommon.TypeInterface).TypeID() != ccurlcommon.CommutativeMeta {
		return errors.New("Error: Wrong variable type, only commutative meta can add a key !")
	}

	if this.Writes() == 0 {
		this.value = this.Value().(ccurlcommon.TypeInterface).Deepcopy()
	}

	meta := this.Value().(*commutative.Meta)
	meta.RefreshCaches(tx, value.(ccurlcommon.UnivalueInterface), source)
	this.IncrementWrite()
	return nil
}

func (this *Univalue) Export(source interface{}) (interface{}, interface{}) {
	if this.Value() != nil {
		this.value = this.value.(ccurlcommon.TypeInterface).Delta(source.(ccurlcommon.LocalCacheInterface).Buffer())
	}

	accessRecord := &Univalue{ // For the arbitrator, just make a deep copy and clear the value field
		tx:          this.GetTx(),
		vType:       this.GetTypeID(this.Value()),
		path:        this.GetPath(),
		reads:       this.Reads(),
		writes:      this.Writes(),
		value:       this.Value(),
		preexists:   this.Preexist(),
		addOrDelete: this.IfAddOrDelete(),
		composite:   this.IfComposite(),
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
	if this.writes == 0 && this.value != nil { // make a deep copy if haven't yet
		this.value = this.value.(ccurlcommon.TypeInterface).Deepcopy()
	}

	this.writes++
	this.tx = tx

	if this.Value() != nil && value != nil && this.vType != value.(ccurlcommon.TypeInterface).TypeID() {
		return errors.New("Error: Types don't match !")
	}

	if this.Value() == nil && value == nil { // Try to delete something nonexistent
		return errors.New("Error: Tried to delete an nonexistent element !")
	}

	if this.Value() == nil && value != nil { // new value
		this.vType = value.(ccurlcommon.TypeInterface).TypeID()
		this.value = value
		this.addOrDelete = this.IfAddOrDelete()
		return nil
	}

	this.writes--
	if r, w, err := this.value.(ccurlcommon.TypeInterface).Set(tx, path, value, source); err == nil { // assignment
		this.writes += w
		this.reads += r
	} else {
		return err
	}

	if value == nil { // Delete an entry
		this.vType = uint8(reflect.Invalid)
		this.value = value
	}
	return nil
}

func (this *Univalue) ApplyDelta(tx uint32, v interface{}) error {
	others := v.([]ccurlcommon.UnivalueInterface)
	for _, other := range others {
		if other == nil {
			continue
		}

		this.PrecheckAttributes(this, other.(*Univalue)) /* Precheck */
		this.writes += other.(*Univalue).writes          /* Merge info */
		this.reads += other.(*Univalue).reads
		this.composite = this.composite && other.(*Univalue).composite
	}

	this.value = this.Value().(ccurlcommon.TypeInterface).ApplyDelta(tx, others)
	return nil
}

func (*Univalue) PrecheckAttributes(this *Univalue, other *Univalue) error {
	if uint8(other.writes) == 0 {
		panic("Error: Value type mismatched!")
	}

	if this.composite != other.composite {
		panic("Error: The composite attribute must match in different transitions")
	}

	if this.addOrDelete != other.addOrDelete {
		this.Print()
		fmt.Println("================================================================")
		other.Print()
		panic("Error: addOrDelete must match")
	}

	if this.preexists != other.preexists {
		this.Print()
		fmt.Println("================================================================")
		other.Print()
		panic("Error: Preexistence must match")
	}

	if this.Value() == nil && this.composite {
		panic("Error: A deleted value cann't be composite")
	}

	if !this.preexists && this.composite {
		panic("Error: A new value cann't be composite")
	}

	return nil
}

func (this *Univalue) PostcheckAttributes() error {
	if this.composite && this.reads > 0 {
		panic("Error: Inconsistent properites")
	}
	return nil
}

func (this *Univalue) Deepcopy() interface{} {
	v := &Univalue{
		vType:       this.vType,
		tx:          this.tx,
		path:        this.path,
		reads:       this.reads,
		writes:      this.writes,
		value:       this.value,
		preexists:   this.preexists,
		addOrDelete: this.addOrDelete,
		composite:   this.composite,
	}

	if v.value != nil {
		v.value = this.value.(ccurlcommon.TypeInterface).Deepcopy()
	}
	return v
}

func (this *Univalue) Print() {
	spaces := fmt.Sprintf("%"+strconv.Itoa(len(strings.Split(this.path, "/"))*4)+"v", " ")
	fmt.Println(spaces+"tx: ", this.tx)
	fmt.Println(spaces+"reads: ", this.reads)
	fmt.Println(spaces+"writes: ", this.writes)
	fmt.Println(spaces+"path: ", this.path)
	fmt.Println(spaces+"value: ", this.value)
	fmt.Println(spaces+"preexists: ", this.preexists)
	fmt.Println(spaces+"addOrDelete: ", this.addOrDelete)
	fmt.Println(spaces+"composite: ", this.composite)
	//this.value.(ccurlcommon.TypeInterface).Print()
	fmt.Println("--------------------------------------------------------")
}

func (this *Univalue) Encode() []byte {
	vBytes := []byte{}
	if this.value != nil {
		vBytes = this.value.(ccurlcommon.TypeInterface).Encode()

	}
	//fmt.Printf("this.path=%v,vBytes=%x,this.vType=%v\n", this.path, vBytes, this.vType)
	return codec.Byteset{
		codec.Uint32(this.vType).Encode(),
		codec.Uint32(uint32(this.tx)).Encode(),
		codec.String(this.path).Encode(),
		codec.Uint32(this.reads).Encode(),
		codec.Uint32(this.writes).Encode(),
		vBytes,
		codec.Bool(this.preexists).Encode(),
		codec.Bool(this.addOrDelete).Encode(),
		codec.Bool(this.composite).Encode(),
	}.Encode()
}

func (this *Univalue) Decode(bytes []byte) interface{} {
	fields := encoding.Byteset{}.Decode(bytes)
	if len(fields) == 0 {
		return nil
	}

	vType := uint8(reflect.Kind(codec.Uint32(0).Decode(fields[0])))
	univalue := &Univalue{
		tx:          uint32(codec.Uint32(0).Decode(fields[1])),
		path:        codec.String("").Decode(fields[2]),
		reads:       uint32(codec.Uint32(0).Decode(fields[3])),
		writes:      uint32(codec.Uint32(0).Decode(fields[4])),
		vType:       vType,
		value:       (&Decoder{}).Decode(fields[5], vType),
		preexists:   bool(codec.Bool(true).Decode(fields[6])),
		addOrDelete: bool(codec.Bool(true).Decode(fields[7])),
		composite:   bool(codec.Bool(true).Decode(fields[8])),
	}

	//fmt.Printf("this.path=%v,vBytes=%x,value=%v\n", univalue.path, fields[5], univalue.Value())
	return univalue
}

func (this *Univalue) GobEncode() ([]byte, error) {
	return this.Encode(), nil
}

func (this *Univalue) GobDecode(data []byte) error {
	v := this.Decode(data).(*Univalue)
	this.vType = v.vType
	this.addOrDelete = v.addOrDelete
	this.composite = v.composite
	this.path = v.path
	this.preexists = v.preexists
	this.reads = v.reads
	this.tx = v.tx
	this.value = v.Value
	this.writes = v.writes
	return nil
}

func (this *Univalue) Equal(other *Univalue) bool {
	if this.vType == other.vType &&
		this.tx == other.tx &&
		this.path == other.path &&
		this.reads == other.reads &&
		this.writes == other.writes &&
		this.value == other.value &&
		this.preexists == other.preexists &&
		this.addOrDelete == other.addOrDelete {
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

	return this.tx == other.tx &&
		this.path == other.path &&
		this.reads == other.reads &&
		this.writes == other.writes &&
		vFlag &&
		this.preexists == other.preexists
}
