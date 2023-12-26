package interfaces

const (
	MAX_DEPTH            uint8 = 12
	ETH10_ACCOUNT_LENGTH       = 40
)

type Platform interface { // value type
	IsSysPath(string) bool
	// Eth10Account() string
}

type Transition interface { // value type
	Added() interface{}
	Removed() interface{}
}

type Type interface { // value type
	TypeID() uint8
	Equal(interface{}) bool
	Clone() interface{}

	IsNumeric() bool
	IsCommutative() bool
	IsBounded() bool

	Value() interface{} // Get() - read/write count
	Delta() interface{}
	DeltaSign() bool
	CloneDelta() interface{}
	Min() interface{}
	Max() interface{}
	New(interface{}, interface{}, interface{}, interface{}, interface{}) interface{}

	SetValue(v interface{})
	IsDeltaApplied() bool
	ResetDelta()
	SetDelta(v interface{})
	SetDeltaSign(v interface{})
	SetMin(v interface{})
	SetMax(v interface{})

	Get() (interface{}, uint32, uint32) // Value, reads and writes, no deltawrites.
	Set(interface{}, interface{}) (interface{}, uint32, uint32, uint32, error)
	CopyTo(interface{}) (interface{}, uint32, uint32, uint32)
	ApplyDelta(interface{}) (Type, int, error)
	IsSelf(interface{}) bool

	MemSize() uint32 // Size in memory
	Size() uint32    // Encoded size
	Encode() []byte
	EncodeToBuffer([]byte) int
	Decode([]byte) interface{}

	StorageEncode() []byte
	StorageDecode([]byte) interface{}

	Hash(func([]byte) []byte) []byte
	Reset()
	Print()
}

type Univalue interface { // value type
	TypeID() uint8
	Reads() uint32
	Writes() uint32
	DeltaWrites() uint32

	From(Univalue) interface{}
	Do(uint32, string, interface{}) interface{}
	Clone() interface{}

	IncrementReads(uint32)
	IncrementWrites(uint32)
	IncrementDeltaWrites(uint32)

	// IsHotLoaded() bool
	Set(uint32, string, interface{}, interface{}) error
	Get(uint32, string, interface{}) interface{}
	GetTx() uint32
	SetTx(uint32)
	GetPath() *string
	SetPath(*string)
	Value() interface{}
	SetValue(interface{}) Univalue

	Merge(WriteCache)
	GetUnimeta() interface{}
	GetCache() interface{}
	New(interface{}, interface{}, interface{}) interface{}

	ApplyDelta(interface{}) error
	Preexist() bool

	Persistent() bool
	IsReadOnly() bool
	IsConcurrentWritable() bool
	// Clone() interface{}
	GetEncoded() []byte
	Size() uint32
	Encode() []byte
	EncodeToBuffer([]byte) int
	Decode([]byte) interface{}
	ClearCache()

	Equal(Univalue) bool
	Print()
}

type WriteCache interface {
	Read(uint32, string, any) (interface{}, interface{})
	Peek(string, any) (interface{}, interface{})
	Write(uint32, string, interface{}) (int64, error)
	AddTransitions([]Univalue)
	// Export() []Univalue

	Export(...func([]Univalue) []Univalue) []Univalue
	// ExportAll( ...func([]ccurlintf.Univalue) []ccurlintf.Univalue) ([]ccurlintf.Univalue, []ccurlintf.Univalue)
	Retrive(string, any) (interface{}, error)
	Cache() *map[string]Univalue
	Store() ReadOnlyDataStore
}

type ReadOnlyDataStore interface {
	IfExists(string) bool
	Retrive(string, any) (interface{}, error)
}

type Datastore interface {
	IfExists(string) bool
	Inject(string, any) error
	BatchInject([]string, []any) error
	Retrive(string, any) (interface{}, error)
	BatchRetrive([]string, []any) []interface{}
	Precommit([]string, interface{}) [32]byte
	Commit(uint64) error
	UpdateCacheStats([]interface{})

	Encoder() func(string, interface{}) []byte
	Decoder() func([]byte, any) interface{}

	// Buffers() ([]string, []interface{}, [][]byte)
	Dump() ([]string, []interface{})
	Clear()
	Print()
	CheckSum() [32]byte
	Query(string, func(string, string) bool) ([]string, [][]byte, error)
	// CacheRetrive(key string, valueTransformer func(interface{}) interface{}) (interface{}, error)
}

type Hasher func(Type) []byte
