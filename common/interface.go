package common

type TypeInterface interface { // value type
	TypeID() uint8
	Deepcopy() interface{}
	Value() interface{}
	Transitional(source interface{}) interface{}
	ToAccess() interface{}
	Get(uint32, string, interface{}) (interface{}, uint32, uint32)
	Set(uint32, string, interface{}, interface{}) (uint32, uint32, error)
	Merge(uint32, interface{})
	Composite() bool
	Finalize()
	Encode() []byte
	Decode([]byte) interface{}
	EncodeStripped() []byte
	DecodeStripped([]byte) interface{}
	Purge()
	Print()
}

type UnivalueInterface interface { // value type
	IsReal() bool
	Set(uint32, string, interface{}, interface{}) error
	Get(uint32, string, interface{}) interface{}
	GetTx() uint32
	GetPath() string
	GetValue() interface{}
	GetReads() uint32
	GetWrites() uint32
	IncrementRead()
	IncrementWrite()
	Merge(uint32, interface{}) error
	Finalize()
	SetPreexist(interface{})
	GetPreexist() bool
	IsComposite() bool
	IsAddOrDelete() bool
	Export(interface{}) (interface{}, interface{}, interface{})
	Encode() []byte
	Print()
}

type DataSourceInterface interface {
	NewValue(uint32, string, interface{}) UnivalueInterface
	Read(uint32, string) interface{}
	Write(uint32, string, interface{}) error
	Insert(string, interface{})
	IfExists(string) bool
	Retrive(string) interface{}
	CheckHistory(uint32, string, bool) UnivalueInterface
	Buffer() *map[string]UnivalueInterface
	Store() *DB
	Commit([]uint32) []error
}

type DB interface {
	Save(string, interface{})
	Retrive(string) interface{}
	BatchSave([]string, []interface{})
	Print()
}

type Decoder interface { // value type
	Decode(bytes []byte, dtype uint8) interface{}
}
