package common

type TypeInterface interface { // value type
	TypeID() uint8
	Deepcopy() interface{}
	Value() interface{}
	Delta() interface{}
	ToAccess() interface{}
	Get(interface{}) (interface{}, uint32, uint32)
	Set(interface{}, interface{}) (uint32, uint32, error)
	This(interface{}) interface{}
	ApplyDelta(interface{}) TypeInterface
	ConcurrentWritable() bool
	IsSelf(interface{}) bool
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
	IncrementReads(uint32)
	IncrementWrites(uint32)
	IncrementDelta(uint32)
	DecrementReads()

	Set(uint32, string, interface{}, interface{}) error
	Get(uint32, string, interface{}) interface{}
	This(interface{}) interface{}
	GetTx() uint32
	GetPath() *string
	SetPath(string)
	Value() interface{}
	SetValue(interface{})

	Reads() uint32
	Writes() uint32
	DeltaWrites() uint32

	GetTransitionType() uint8
	SetTransitionType(uint8)
	ApplyDelta(uint32, interface{}) error
	Preexist() bool
	ConcurrentWritable() bool // Delta writable
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
	Peek(path string) (interface{}, bool)
	Write(uint32, string, interface{}) error
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

type Hasher func(TypeInterface) []byte
