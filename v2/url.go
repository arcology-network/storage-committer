package concurrenturl

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"sort"
	"time"

	"github.com/arcology-network/common-lib/common"
	"github.com/holiman/uint256"

	// performance "github.com/arcology-network/common-lib/mhasher"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	indexer "github.com/arcology-network/concurrenturl/v2/indexer"
	commutative "github.com/arcology-network/concurrenturl/v2/type/commutative"
	"github.com/arcology-network/concurrenturl/v2/type/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/v2/univalue"
)

type ConcurrentUrl struct {
	indexer    *indexer.Indexer
	invIndexer *indexer.Indexer

	Platform *ccurlcommon.Platform
	// Buf for Export.
	buffer        []ccurlcommon.UnivalueInterface // Transition + access record buffer
	accesseBuf    []ccurlcommon.UnivalueInterface // Access records
	transitBuf    []ccurlcommon.UnivalueInterface // Transitions
	ImportFilters []ccurlcommon.FilterTransitionsInterface
	numThreads    int
}

func NewConcurrentUrl(store ccurlcommon.DatastoreInterface, args ...interface{}) *ConcurrentUrl {
	platform := ccurlcommon.NewPlatform()
	return &ConcurrentUrl{
		indexer:    indexer.NewIndexer(store, platform),
		invIndexer: indexer.NewIndexer(store, platform),
		Platform:   platform,

		buffer:     make([]ccurlcommon.UnivalueInterface, 0, 64),
		accesseBuf: make([]ccurlcommon.UnivalueInterface, 0, 64),
		transitBuf: make([]ccurlcommon.UnivalueInterface, 0, 64),

		ImportFilters: []ccurlcommon.FilterTransitionsInterface{&indexer.NonceFilter{}, &indexer.BalanceFilter{}},
		numThreads:    8,
	}
}

// Get data from the DB direcly, still under conflict protection
func (this *ConcurrentUrl) ReadCommitted(tx uint32, key string) (interface{}, error) {
	if _, err := this.Read(tx, key); err != nil { // For conflict detection
		return nil, err
	}
	return (*this.Store()).Retrive(key)
}

func (this *ConcurrentUrl) Indexer() *indexer.Indexer {
	return this.indexer
}

func (this *ConcurrentUrl) Init(store ccurlcommon.DatastoreInterface) {
	this.indexer.Init(store)
	this.invIndexer.Init(store)
	this.reset()
}

func (this *ConcurrentUrl) reset() {
	this.buffer = this.buffer[:0]
	this.accesseBuf = this.accesseBuf[:0]
	this.transitBuf = this.transitBuf[:0]
}

func (this *ConcurrentUrl) Store() *ccurlcommon.DatastoreInterface {
	return this.indexer.Store()
}

func (this *ConcurrentUrl) Clear() {
	(*this.indexer.Store()).Clear()
	this.reset() // Reset the buffers
	this.indexer.Clear()
	this.invIndexer.Clear()
}

// load accounts
func (this *ConcurrentUrl) CreateAccount(tx uint32, platform string, acct string) error {
	paths, typeids := this.Platform.GetBuiltins(acct)

	var err error
	for i, path := range paths {
		var v interface{}
		switch typeids[i] {
		case ccurlcommon.CommutativeMeta: // Path
			v = commutative.NewPath()

		case uint8(reflect.Kind(ccurlcommon.NoncommutativeString)): // delta big int
			v = noncommutative.NewString("")

		case uint8(reflect.Kind(ccurlcommon.CommutativeUint256)): // delta big int
			v = commutative.NewU256(uint256.NewInt(0), commutative.U256MIN, commutative.U256MAX)

		case uint8(reflect.Kind(ccurlcommon.CommutativeUint64)):
			v = commutative.NewUint64(0, math.MaxUint64)

		case uint8(reflect.Kind(ccurlcommon.NoncommutativeInt64)):
			v = noncommutative.NewInt64(0)

		case uint8(reflect.Kind(ccurlcommon.NoncommutativeBytes)):
			v = noncommutative.NewBytes([]byte{})
		}

		if !this.indexer.IfExists(path) {
			err = this.indexer.Write(tx, path, v) // root path
		}
	}
	return err
}

func (this *ConcurrentUrl) IfExists(path string) bool {
	return this.indexer.IfExists(path)
}

func (this *ConcurrentUrl) Peek(path string) (interface{}, error) {
	value, _ := this.indexer.Peek(path)
	return value, nil
}

func (this *ConcurrentUrl) Read(tx uint32, path string) (interface{}, error) {
	return this.indexer.Read(tx, path), nil
}

func (this *ConcurrentUrl) Write(tx uint32, path string, value interface{}) error {
	if value != nil {
		if id := (&univalue.Univalue{}).GetTypeID(value); id == uint8(reflect.Invalid) {
			return errors.New("Error: Unknown data type !")
		}
	}
	return this.indexer.Write(tx, path, value)
}

// Read th Nth element under a path
func (this *ConcurrentUrl) at(tx uint32, path string, idx uint64) (interface{}, error) {
	if !ccurlcommon.IsPath(path) {
		return nil, errors.New("Error: Not a path!!!")
	}

	meta, err := this.Read(tx, path) // read the container meta
	if err != nil {
		return nil, err
	}

	key, ok := meta.(*commutative.Path).View().KeyOf(idx)
	if key == nil || !ok {
		return nil, errors.New("Error: The element wasn't found!!!")
	}

	return path + key.(string), nil
}

// Read th Nth element under a path
func (this *ConcurrentUrl) ReadAt(tx uint32, path string, idx uint64) (interface{}, error) {
	if key, err := this.at(tx, path, idx); err == nil {
		return this.Read(tx, key.(string))
	} else {
		return key, err
	}
}

// Read th Nth element under a path
func (this *ConcurrentUrl) PopBack(tx uint32, path string) (interface{}, error) {
	if !ccurlcommon.IsPath(path) {
		return nil, errors.New("Error: Not a path!!!")
	}

	meta, err := this.Read(tx, path) // read the container meta
	if err != nil {
		return nil, err
	}

	if length := meta.(*commutative.Path).View().Len(); length > 0 {
		if key, ok := meta.(*commutative.Path).View().KeyOf(length - 1); ok {
			value, err := this.Read(tx, path+key.(string))
			if value == nil || err != nil {
				return value, err
			}
			return value, this.Write(tx, path+key.(string), nil)
		}
	}
	return nil, errors.New("Error: Empty container!")
}

// Read th Nth element under a path
func (this *ConcurrentUrl) WriteAt(tx uint32, path string, idx uint64, value interface{}) error {
	if !ccurlcommon.IsPath(path) {
		return errors.New("Error: Not a path!!!")
	}

	if key, err := this.at(tx, path, idx); err == nil {
		return this.Write(tx, key.(string), value)
	} else {
		return err
	}
}

func (this *ConcurrentUrl) Exempted(transition ccurlcommon.UnivalueInterface) bool {
	for i := 0; i < len(this.ImportFilters); i++ {
		if this.ImportFilters[i].Is(this.Platform.RootLength(), *transition.GetPath()) {
			return true
		}
	}
	return false
}

func (this *ConcurrentUrl) Import(transitions []ccurlcommon.UnivalueInterface, args ...interface{}) {
	invTransitions := make([]ccurlcommon.UnivalueInterface, 0, len(transitions))
	for i := 0; i < len(transitions); i++ {
		if this.Exempted(transitions[i]) {
			invTransitions = append(invTransitions, transitions[i])
			transitions[i] = nil
		}
	}
	common.RemoveIf(&transitions, func(v ccurlcommon.UnivalueInterface) bool { return v == nil })

	common.ParallelExecute(
		func() { this.invIndexer.Import(invTransitions, args...) },
		func() { this.indexer.Import(transitions, args...) })

}

func (this *ConcurrentUrl) KVs() ([]string, []interface{}) {
	keys, values := this.indexer.KVs()
	invKeys, invVals := this.invIndexer.KVs()

	kvs := make(map[string]interface{}, len(keys)+len(invKeys))
	for i, key := range keys {
		kvs[key] = values[i]
	}
	for i, key := range invKeys {
		kvs[key] = invVals[i]
	}

	// sortedKeys, err := performance.SortStrings(append(keys, invKeys...)) // Keys should be unique
	// if err != nil {
	// 	panic(err)
	// }
	sortedKeys := append(keys, invKeys...)
	sort.Strings(sortedKeys)

	sortedVals := make([]interface{}, len(sortedKeys))
	sorter := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			sortedVals[i] = kvs[sortedKeys[i]]
		}
	}
	common.ParallelWorker(len(sortedKeys), this.numThreads, sorter)

	return sortedKeys, sortedVals
}

// Call this as s
func (this *ConcurrentUrl) PostImport() {
	common.ParallelExecute(
		func() { this.invIndexer.SortTransitions() },
		func() { this.indexer.SortTransitions() })
}

func (this *ConcurrentUrl) Precommit(txs []uint32) []error {
	if txs != nil && len(txs) == 0 { // Commit all the transactions when txs == nil
		return []error{}
	}

	common.ParallelExecute(
		func() { this.invIndexer.FinalizeStates() },
		func() {
			this.indexer.WhilteList(txs)  // Remove all the transitions generated by the conflicting transactions
			this.indexer.FinalizeStates() // Finalize states
		},
	)
	return nil
}

func (this *ConcurrentUrl) Postcommit() {
	keys, values := this.KVs()
	(*this.indexer.Store()).Precommit(keys, values) // save the transitions to the DB buffer
}

func (this *ConcurrentUrl) SaveToDB() {
	store := this.indexer.Store()
	(*store).Commit() // Commit to the state store
	this.Clear()
}

func (this *ConcurrentUrl) Commit(txs []uint32) []error {
	if len(txs) == 0 {
		this.Clear()
		return nil
	}
	errs := this.Precommit(txs)
	this.Postcommit()
	this.SaveToDB()
	return errs
}

func (this *ConcurrentUrl) AllInOneCommit(transitions []ccurlcommon.UnivalueInterface, txs []uint32) []error {
	t0 := time.Now()

	accountMerkle := indexer.NewAccountMerkle(this.Platform)
	common.ParallelExecute(
		func() { this.indexer.Import(transitions) },
		func() { accountMerkle.Import(transitions) })

	fmt.Println("indexer.Import + accountMerkle Import :--------------------------------", time.Since(t0))

	t0 = time.Now()
	this.PostImport()
	fmt.Println("indexer.Commit :--------------------------------", time.Since(t0))

	t0 = time.Now()
	runtime.GC()
	fmt.Println("GC 0:--------------------------------", time.Since(t0))

	t0 = time.Now()
	this.Precommit(txs)
	fmt.Println("Precommit :--------------------------------", time.Since(t0))

	// Build the merkle tree
	t0 = time.Now()
	k, v := this.indexer.KVs()
	encoded := make([][]byte, 0, len(v))
	for _, value := range v {
		encoded = append(encoded, value.(ccurlcommon.UnivalueInterface).GetEncoded())
	}
	accountMerkle.Build(k, encoded)
	fmt.Println("ComputeMerkle:", time.Since(t0))

	t0 = time.Now()
	this.Postcommit()
	fmt.Println("Postcommit :--------------------------------", time.Since(t0))

	t0 = time.Now()
	this.SaveToDB()
	fmt.Println("SaveToDB :--------------------------------", time.Since(t0))

	return []error{}
}

func (this *ConcurrentUrl) Export(sorter func([]ccurlcommon.UnivalueInterface) interface{}) ([]ccurlcommon.UnivalueInterface, []ccurlcommon.UnivalueInterface) {
	this.indexer.Vectorize(this.indexer.Buffer(), &this.buffer, false) // Export records
	if sorter != nil {                                                 // Sort by path, debug only
		indexer.Sorter(this.buffer)
	}

	this.accesseBuf = indexer.FilterAccesses(this.buffer, this.Platform)
	this.transitBuf = indexer.FilterTransitions(this.buffer, this.Platform)
	return this.accesseBuf, this.transitBuf
}

func (this *ConcurrentUrl) Export2(sorter func([]ccurlcommon.UnivalueInterface) interface{}) ([]ccurlcommon.UnivalueInterface, []ccurlcommon.UnivalueInterface) {
	this.indexer.Vectorize(this.indexer.Buffer(), &this.buffer, false) // Export records
	if sorter != nil {                                                 // Sort by path, debug only
		indexer.Sorter(this.buffer)
	}

	for i := 0; i < len(this.buffer); i++ {
		if this.buffer[i].Writes() == 0 && this.buffer[i].DeltaWrites() == 0 {
			this.buffer[i] = nil // Clear read-only records
		}
	}

	// this.accesseBuf = indexer.FilterAccesses(this.buffer, this.Platform)
	// this.transitBuf = indexer.FilterTransitions(this.buffer, this.Platform)
	return this.accesseBuf, this.transitBuf
}

type PostProcessFunc func(accesses, transitions []ccurlcommon.UnivalueInterface) ([]ccurlcommon.UnivalueInterface, []ccurlcommon.UnivalueInterface)

func (this *ConcurrentUrl) ExportEncoded(ppf PostProcessFunc) ([][]byte, [][]byte) {
	records, transitions := this.Export(nil)
	if ppf != nil {
		records, transitions = ppf(records, transitions)
	}

	recordsBytes := make([][]byte, len(records))
	for i := 0; i < len(records); i++ {
		recordsBytes[i] = records[i].Encode()
	}

	transBytes := make([][]byte, len(transitions))
	for i := 0; i < len(transitions); i++ {
		transBytes[i] = transitions[i].Encode()
	}

	return recordsBytes, transBytes
}

func (this *ConcurrentUrl) Print() {
	this.indexer.Print()
}
