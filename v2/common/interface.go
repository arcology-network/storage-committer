package common

type TypeInterface interface { // value type
	TypeID() uint8
	Deepcopy() interface{}
	Value() interface{}
	Delta(source interface{}) interface{}
	ToAccess() interface{}
	Get(uint32, string, interface{}) (interface{}, uint32, uint32)
	Set(uint32, string, interface{}, interface{}) (uint32, uint32, error)
	Peek(interface{}) interface{}
	ApplyDelta(uint32, interface{}) TypeInterface
	Composite() bool
	Hash(func([]byte) []byte) []byte
	Encode() []byte
	Decode([]byte) interface{}
	EncodeCompact() []byte
	DecodeCompact([]byte) interface{}
	Purge()
	Print()
}

type UnivalueInterface interface { // value type
	Set(uint32, string, interface{}, interface{}) error
	Get(uint32, string, interface{}) interface{}
	UpdateParentMeta(uint32, interface{}, interface{}) error
	Peek(interface{}) interface{}
	GetTx() uint32
	GetPath() string
	Value() interface{}
	Reads() uint32
	Writes() uint32
	ApplyDelta(uint32, interface{}) error
	Preexist() bool
	Composite() bool
	Deepcopy() interface{}
	Export(interface{}) (interface{}, interface{})
	GetCachedEncoded() []byte
	Encode() []byte
	Decode([]byte) interface{}
	Print()
}

type LocalCacheInterface interface {
	NewValue(uint32, string, interface{}) UnivalueInterface
	Read(uint32, string) interface{}
	TryRead(tx uint32, path string) interface{}
	Write(uint32, string, interface{}) error
	Insert(string, interface{})
	IfExists(string) bool
	RetriveShallow(string) interface{}
	CheckHistory(uint32, string, bool) UnivalueInterface
	Buffer() *map[string]UnivalueInterface
	Store() *DB
	Commit([]uint32) ([]string, interface{}, []error)
}

type DB interface {
	Save(string, interface{})
	Retrive(string) interface{}
	BatchSave([]string, interface{})
	Clear()
	Print()
}

type Decoder interface { // value type
	Decode(bytes []byte, dtype uint8) interface{}
}

type Hasher func(TypeInterface) []byte
