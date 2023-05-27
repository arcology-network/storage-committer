package common

import (
	"math"
)

const (
	MAX_DEPTH            uint8 = 12
	SYSTEM                     = math.MaxInt32
	ETH10_ACCOUNT_LENGTH       = 40
)

type PlatformInterface interface { // value type
	IsSysPath(string) bool
	Eth10Account() string
}

type TransitionInterface interface { // value type
	Added() interface{}
	Removed() interface{}
}

type TypeInterface interface { // value type
	TypeID() uint8
	Equal(interface{}) bool
	Clone() interface{}

	IsNumeric() bool
	IsCommutative() bool

	Value() interface{} // Get() - read/write count
	Delta() interface{}
	DeltaSign() bool
	Min() interface{}
	Max() interface{}
	New(interface{}, interface{}, interface{}, interface{}, interface{}) interface{}
	ReInit()

	Get() (interface{}, uint32, uint32)
	Set(interface{}, interface{}) (interface{}, uint32, uint32, uint32, error)
	CopyTo(interface{}) (interface{}, uint32, uint32, uint32)
	ApplyDelta(interface{}) (TypeInterface, int, error)
	IsSelf(interface{}) bool

	MemSize() uint32 // Size in memory
	Size() uint32    // Encoded size
	Encode() []byte
	EncodeToBuffer([]byte) int
	Decode([]byte) interface{}

	Hash(func([]byte) []byte) []byte
	Reset()
	Print()
}

type UnivalueInterface interface { // value type
	TypeID() uint8
	Reads() uint32
	Writes() uint32
	DeltaWrites() uint32

	Do(uint32, string, interface{}) interface{}
	Clone() interface{}
	// Meta() UnivalueInterface

	// Is(string) bool

	IncrementReads(uint32)
	IncrementWrites(uint32)
	IncrementDelta(uint32)

	GetErrorCode() uint8
	SetErrorCode(uint8)

	IsHotLoaded() bool
	Set(uint32, string, interface{}, interface{}) error
	Get(uint32, string, interface{}) interface{}
	GetTx() uint32
	SetTx(uint32)
	GetPath() *string
	SetPath(*string)
	Value() interface{}
	SetValue(interface{})

	WriteTo(WriteCacheInterface)
	GetUnimeta() interface{}
	GetCache() interface{}
	New(interface{}, interface{}, interface{}, interface{}) interface{}

	ApplyDelta(interface{}) error
	Preexist() bool
	IsReadOnly() bool
	IsConcurrentWritable() bool
	// Clone() interface{}

	GetEncoded() []byte
	Size() uint32
	Encode() []byte
	EncodeToBuffer([]byte) int
	Decode([]byte) interface{}
	ClearCache()

	Equal(UnivalueInterface) bool
	Print()
}

type WriteCacheInterface interface {
	Read(uint32, string) (interface{}, interface{})
	Peek(path string) (interface{}, interface{})
	Write(uint32, string, interface{}) error
	Insert(string, interface{})

	RetriveShallow(string) interface{}
	Cache() *map[string]UnivalueInterface
	Store() DatastoreInterface
}

type ImporterInterface interface {
	RetriveShallow(string) interface{}
}

type FilteredTransitionsInterface interface {
	Is(int, string) bool
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
