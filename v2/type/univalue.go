package urltype

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	codec "github.com/HPISTechnologies/common-lib/codec"
	encoding "github.com/HPISTechnologies/common-lib/encoding"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
	commutative "github.com/HPISTechnologies/concurrenturl/v2/type/commutative"
	noncommutative "github.com/HPISTechnologies/concurrenturl/v2/type/noncommutative"
)

type Univalue struct {
	vType       uint8
	Tx          uint32
	Path        string
	Reads       uint32
	Writes      uint32
	Value       interface{}
	Preexists   bool
	AddOrDelete bool
	Composite   bool
}

func NewUnivalue(tx uint32, key string, Reads, Writes uint32, args ...interface{}) *Univalue {
	v := &Univalue{
		vType:       (&Univalue{}).GetTypeID(args[0]),
		Tx:          tx,
		Path:        key,
		Reads:       Reads,
		Writes:      Writes,
		Value:       args[0],
		Preexists:   false,
		AddOrDelete: false,
		Composite:   false,
	}

	if len(args) > 1 {
		v.SetPreexist(args[1])
		v.AddOrDelete = v.IsAddOrDelete()
		v.Composite = v.IsComposite()
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

func (this *Univalue) GetTx() uint32         { return this.Tx }
func (this *Univalue) GetPath() string       { return this.Path }
func (this *Univalue) GetValue() interface{} { return this.Value }
func (this *Univalue) GetReads() uint32      { return this.Reads }
func (this *Univalue) GetWrites() uint32     { return this.Writes } // Exist in cache as a failed read
func (this *Univalue) IncrementRead()        { this.Reads++ }
func (this *Univalue) IncrementWrite()       { this.Writes++ }
func (this *Univalue) GetPreexist() bool     { return this.Preexists } // Exist in cache as a failed read
func (this *Univalue) SetPreexist(source interface{}) {
	this.Preexists = source.(ccurlcommon.DataSourceInterface).Retrive(this.Path) != nil
}
func (this *Univalue) IsReal() bool { return this.Preexists || this.Writes > 0 } // A real element or just an access record
func (this *Univalue) IsAddOrDelete() bool {
	return (!this.GetPreexist() && this.GetWrites() > 0) || (this.GetPreexist() && this.GetValue() == nil)
}
func (this *Univalue) IsComposite() bool { // Call this before setting the value attribute to nil
	if this.Value != nil && this.GetPreexist() {
		return this.Value.(ccurlcommon.TypeInterface).Composite() && this.Reads == 0
	}
	return false
}

func (this *Univalue) Export(source interface{}) (interface{}, interface{}, interface{}) {
	if this.GetValue() != nil {
		this.Value = this.Value.(ccurlcommon.TypeInterface).Delta(source.(ccurlcommon.DataSourceInterface).Buffer())
	}

	accessRecord := &Univalue{ // For the arbitrator, just make a deep copy and clear the value field
		Tx:          this.GetTx(),
		vType:       this.GetTypeID(this.GetValue()),
		Path:        this.GetPath(),
		Reads:       this.GetReads(),
		Writes:      this.GetWrites(),
		Value:       this.GetValue(),
		Preexists:   this.GetPreexist(),
		AddOrDelete: this.IsAddOrDelete(),
		Composite:   this.IsComposite(),
	}

	if accessRecord.GetValue() != nil {
		accessRecord.Value = accessRecord.GetValue().(ccurlcommon.TypeInterface).ToAccess()
	}

	if this.GetWrites() > 0 && this.GetValue() != nil { // Rewrite an existing entry or create a new one
		return accessRecord, this, this.GenerateAuxTran(source)
	}

	if this.GetWrites() > 0 && this.GetValue() == nil && this.GetPreexist() { // Deletion of an existing entry
		return accessRecord, this, this.GenerateAuxTran(source)
	}

	return accessRecord, nil, this.GenerateAuxTran(source)
}

func (this *Univalue) GenerateAuxTran(source interface{}) interface{} {
	if this.IsAddOrDelete() {
		parentPath := ccurlcommon.GetParentPath(this.GetPath())
		auxTran := source.(ccurlcommon.DataSourceInterface).CheckHistory(this.Tx, parentPath, false).(*Univalue)
		auxTran.Composite = auxTran.IsComposite()
		auxTran.IncrementWrite()
		return auxTran
	}
	return nil
}

func (this *Univalue) Get(tx uint32, path string, source interface{}) interface{} {
	if this.Value != nil {
		tempV, r, w := this.Value.(ccurlcommon.TypeInterface).Get(tx, path, source) //RW: Affiliated reads and writes
		this.Reads += r
		this.Writes += w
		this.Composite = tempV.(ccurlcommon.TypeInterface).Composite()
		return tempV
	}
	this.IncrementRead()
	return this.Value
}

func (this *Univalue) Peek(source interface{}) interface{} {
	if this.Value != nil {
		return this.Value.(ccurlcommon.TypeInterface).Peek(source)
	}
	return this.Value
}

func (this *Univalue) Set(tx uint32, path string, value interface{}, source interface{}) error { // update the value
	this.Writes++
	this.Tx = tx

	if this.GetValue() != nil && value != nil && this.vType != value.(ccurlcommon.TypeInterface).TypeID() {
		return errors.New("Error: Types don't match !")
	}

	if this.GetValue() == nil && value == nil {
		return errors.New("Error: Tried to delete an nonexistent element !")
	}

	if this.GetValue() == nil && value != nil { // assignment
		this.vType = value.(ccurlcommon.TypeInterface).TypeID()
		this.Value = value
		this.AddOrDelete = this.IsAddOrDelete()
		return nil
	}

	this.Writes--
	if r, w, err := this.Value.(ccurlcommon.TypeInterface).Set(tx, path, value, source); err == nil {
		this.Writes += w
		this.Reads += r
	} else {
		return err
	}

	if value == nil { // Delete an entry
		this.vType = uint8(reflect.Invalid)
		this.Value = value
	}
	return nil
}

func (this *Univalue) ApplyDelta(tx uint32, other interface{}) error {
	if other.(*Univalue).GetValue() == nil {
		this.Value = nil
	} else {
		this.PrecheckAttributes(this, other.(*Univalue)) /* Precheck */
		this.Writes += other.(*Univalue).Writes          /* Merge info */
		this.Reads += other.(*Univalue).Reads
		this.Composite = this.Composite && other.(*Univalue).Composite
		this.GetValue().(ccurlcommon.TypeInterface).ApplyDelta(tx, other.(*Univalue).GetValue())
	}
	return nil
}

func (*Univalue) PrecheckAttributes(this *Univalue, other *Univalue) error {
	if this.GetValue().(ccurlcommon.TypeInterface).TypeID() != other.GetValue().(ccurlcommon.TypeInterface).TypeID() {
		panic("Error: Value type mismatched!")
	}

	if this.Tx == other.Tx && this.Composite != other.Composite {
		panic("Error: The composite attribute must match in different transitions")
	}

	if this.AddOrDelete != other.AddOrDelete {
		panic("Error: AddOrDelete must match")
	}

	if this.Preexists != other.Preexists {
		panic("Error: Preexistence must match")
	}

	if this.GetValue() == nil && this.Composite {
		panic("Error: A deleted value cann't be composite")
	}

	if !this.Preexists && this.Composite {
		panic("Error: A new value cann't be composite")
	}

	return nil
}

func (this *Univalue) PostcheckAttributes() error {
	if this.Composite && this.Reads > 0 {
		panic("Error: Inconsistent properites")
	}
	return nil
}

func (this *Univalue) Finalize() {
	if this.Value == nil {
		return
	}
	this.Value.(ccurlcommon.TypeInterface).Finalize()
}

func (this *Univalue) Deepcopy() interface{} {
	v := *this
	v.Value = this.GetValue().(ccurlcommon.TypeInterface).Deepcopy()
	return v
}

func (this *Univalue) Print() {
	spaces := fmt.Sprintf("%"+strconv.Itoa(len(strings.Split(this.Path, "/"))*4)+"v", " ")
	fmt.Println(spaces+"tx: ", this.Tx)
	fmt.Println(spaces+"Reads: ", this.Reads)
	fmt.Println(spaces+"Writes: ", this.Writes)
	fmt.Println(spaces+"path: ", this.Path)
	fmt.Println(spaces+"value: ", this.Value)
	fmt.Println(spaces+"Preexists: ", this.Preexists)
	fmt.Println(spaces+"addOrDelete: ", this.AddOrDelete)
	fmt.Println(spaces+"Composite: ", this.Composite)
	//this.Value.(ccurlcommon.TypeInterface).Print()
	fmt.Println("--------------------------------------------------------")
}

func (this *Univalue) Encode() []byte {
	vBytes := []byte{}
	if this.Value != nil {
		vBytes = this.Value.(ccurlcommon.TypeInterface).Encode()
	}

	return codec.Byteset{
		codec.Uint32(this.vType).Encode(),
		codec.Uint32(uint32(this.Tx)).Encode(),
		codec.String(this.Path).Encode(),
		codec.Uint32(this.Reads).Encode(),
		codec.Uint32(this.Writes).Encode(),
		vBytes,
		codec.Bool(this.Preexists).Encode(),
		codec.Bool(this.AddOrDelete).Encode(),
		codec.Bool(this.Composite).Encode(),
	}.Encode()
}

func (this *Univalue) Decode(bytes []byte) interface{} {
	fields := encoding.Byteset{}.Decode(bytes)
	if len(fields) == 0 {
		return nil
	}

	vType := uint8(reflect.Kind(codec.Uint32(0).Decode(fields[0])))
	univalue := &Univalue{
		Tx:          uint32(codec.Uint32(0).Decode(fields[1])),
		Path:        codec.String("").Decode(fields[2]),
		Reads:       uint32(codec.Uint32(0).Decode(fields[3])),
		Writes:      uint32(codec.Uint32(0).Decode(fields[4])),
		vType:       vType,
		Value:       (&Decoder{}).Decode(fields[5], vType),
		Preexists:   bool(codec.Bool(true).Decode(fields[6])),
		AddOrDelete: bool(codec.Bool(true).Decode(fields[7])),
		Composite:   bool(codec.Bool(true).Decode(fields[8])),
	}
	return univalue
}

func (this *Univalue) Equal(other *Univalue) bool {
	if this.vType == other.vType &&
		this.Tx == other.Tx &&
		this.Path == other.Path &&
		this.Reads == other.Reads &&
		this.Writes == other.Writes &&
		this.Value == other.Value &&
		this.Preexists == other.Preexists &&
		this.AddOrDelete == other.AddOrDelete {
		return true
	}
	return false
}

func (this *Univalue) EqualAccess(other *Univalue) bool {
	if this.Tx == other.Tx &&
		this.Path == other.Path &&
		this.Reads == other.Reads &&
		this.Writes == other.Writes &&
		reflect.DeepEqual(this.Value, other.Value) &&
		this.Preexists == other.Preexists {
		return true
	}
	return false
}
