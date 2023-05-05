package common

import (
	"bytes"
	"math"
	"sort"
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

	Value() interface{} // Get() - read/write count
	Delta() interface{}
	Sign() bool
	Min() interface{}
	Max() interface{}
	New(interface{}, interface{}, interface{}, interface{}, interface{}) interface{}
	ReInit()

	Get() (interface{}, uint32, uint32)
	Set(interface{}, interface{}) (interface{}, uint32, uint32, uint32, error)
	CopyTo(interface{}) (interface{}, uint32, uint32, uint32)
	ApplyDelta(interface{}) TypeInterface
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

	Delta() UnivalueInterface
	Meta() UnivalueInterface

	IncrementReads(uint32)
	IncrementWrites(uint32)
	IncrementDelta(uint32)

	Set(uint32, string, interface{}, interface{}) error
	Get(uint32, string, interface{}) interface{}
	GetTx() uint32
	GetPath() *string
	SetPath(*string)
	Value() interface{}
	SetValue(interface{})

	GetUnimeta() interface{}
	GetCache() interface{}
	New(interface{}, interface{}, interface{}) interface{}

	ApplyDelta(uint32, interface{}) error
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

type IndexerInterface interface {
	Read(uint32, string) interface{}
	Peek(path string) (interface{}, bool)
	Write(uint32, string, interface{}) error
	Insert(string, interface{})

	RetriveShallow(string) interface{}
	Buffer() *map[string]UnivalueInterface
	Store() *DatastoreInterface

	//Platform() *Platform
}

type FilterTransitionsInterface interface {
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

func Sorter(univals []UnivalueInterface) []UnivalueInterface {
	sort.SliceStable(univals, func(i, j int) bool {
		lhs := (*(univals[i].GetPath()))
		rhs := (*(univals[j].GetPath()))
		return bytes.Compare([]byte(lhs)[:], []byte(rhs)[:]) < 0
	})
	return univals
}
