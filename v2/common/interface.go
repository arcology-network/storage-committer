package common

type TypeInterface interface { // value type
	TypeID() uint8
	Deepcopy() interface{}
	Value() interface{}
	Delta(source interface{}) interface{}
	ToAccess() interface{}
	Get(string, interface{}) (interface{}, uint32, uint32)
	Set(string, interface{}, interface{}) (uint32, uint32, error)
	Reset(string, interface{}, interface{}) (uint32, uint32, error)
	This(interface{}) interface{}
	ApplyDelta(interface{}) TypeInterface
	Composite() bool
	Hash(func([]byte) []byte) []byte
	Encode() []byte
	EncodeToBuffer([]byte) int
	Decode([]byte) interface{}
	Size() uint32
	EncodeCompact() []byte
	DecodeCompact([]byte) interface{}
	Purge()
	Print()
}

type UnivalueInterface interface { // value type
	DecrementReads()
	Set(uint32, string, interface{}, interface{}) error
	Reset(uint32, string, interface{}, interface{}) error
	Get(uint32, string, interface{}) interface{}
	UpdateParentMeta(uint32, interface{}, interface{}) bool
	This(interface{}) interface{}
	GetTx() uint32
	GetPath() *string
	SetPath(string)
	Value() interface{}
	SetValue(interface{})
	Reads() uint32
	Writes() uint32
	GetTransitionType() uint8
	SetTransitionType(uint8)
	ApplyDelta(uint32, interface{}) error
	Preexist() bool
	Composite() bool
	Deepcopy() interface{}
	Export(interface{}) (interface{}, interface{})
	GetEncoded() []byte
	Encode() []byte
	Decode([]byte) interface{}
	ClearReserve()
	Print()
}

type IndexerInterface interface {
	Read(uint32, string) interface{}
	TryRead(tx uint32, path string) (interface{}, bool)
	Write(uint32, string, interface{}, bool) error
	Insert(string, interface{})

	RetriveShallow(string) interface{}
	Buffer() *map[string]UnivalueInterface
	Store() *DatastoreInterface

	SkipExportTransitions(univalue interface{}) bool
}

type DatastoreInterface interface {
	Inject(string, interface{})
	BatchInject([]string, []interface{})
	Retrive(string) (interface{}, error)
	BatchRetrive([]string) []interface{}
	Precommit([]string, interface{})
	Commit() error
	UpdateCacheStats([]interface{})
	Dump() ([]string, []interface{})
	Checksum() [32]byte
	Clear()
	Print()
	CheckSum() [32]byte
	Query(string, func(string, string) bool) ([]string, [][]byte, error)
	CacheRetrive(key string, valueTransformer func(interface{}) interface{}) (interface{}, error)
}

// type DecoderInterface interface { // value type
// 	Decode(bytes []byte, dtype uint8) interface{}
// }

type Hasher func(TypeInterface) []byte
