package interfaces

const (
	MAX_DEPTH            uint8 = 12
	ETH10_ACCOUNT_LENGTH       = 40
)

type Platform interface { // value type
	IsSysPath(string) bool
	// Eth10Account() string
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
	ApplyDelta([]Type) (Type, int, error)
	IsSelf(interface{}) bool

	MemSize() uint32 // Size in memory
	Size() uint32    // Encoded size
	Encode() []byte
	EncodeToBuffer([]byte) int
	Decode([]byte) interface{}

	StorageEncode(string) []byte
	StorageDecode(string, []byte) interface{}

	Preload(string, interface{})

	Hash(func([]byte) []byte) []byte
	Reset()
	Print()
}

type WritableCacheInterface interface {
	IfExists(string) bool
	Write(uint32, string, interface{}) (int64, error)
}

type Indexer[T any] interface {
	Add([]T)
	Get() any
	Finalize()
	Clear()
}

type ReadOnlyDataStore interface {
	IfExists(string) bool
	Retrive(string, any) (interface{}, error)
}

type CommittableStore interface {
	Precommit(...interface{}) [32]byte //key gatter, value gatter,
	Commit(uint64) error
}

type Datastore interface {
	ReadOnlyDataStore
	CommittableStore
	Inject(string, any) error
	Preload([]byte) interface{}
}

type Hasher func(Type) []byte
