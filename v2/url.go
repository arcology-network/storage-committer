package concurrenturl

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"sort"
	"time"

	"github.com/arcology-network/common-lib/common"
	// performance "github.com/arcology-network/common-lib/mhasher"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	ccurltype "github.com/arcology-network/concurrenturl/v2/type"
	commutative "github.com/arcology-network/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/v2/type/noncommutative"
	"github.com/holiman/uint256"
)

type ConcurrentUrl struct {
	indexer    *ccurltype.Indexer
	invIndexer *ccurltype.Indexer

	Platform *ccurlcommon.Platform
	// Buf for Export.
	records    []ccurlcommon.UnivalueInterface // Transition + access record buffer
	accesseBuf []ccurlcommon.UnivalueInterface // Access records
	transitBuf []ccurlcommon.UnivalueInterface // Transitions

	numThreads int
}

func NewConcurrentUrl(store ccurlcommon.DatastoreInterface, args ...interface{}) *ConcurrentUrl {
	platform := ccurlcommon.NewPlatform()
	return &ConcurrentUrl{
		indexer:    ccurltype.NewIndexer(store, platform),
		invIndexer: ccurltype.NewIndexer(store, platform),
		Platform:   platform,

		records:    make([]ccurlcommon.UnivalueInterface, 0, 64),
		accesseBuf: make([]ccurlcommon.UnivalueInterface, 0, 64),
		transitBuf: make([]ccurlcommon.UnivalueInterface, 0, 64),

		numThreads: 8,
	}
}

// Get data from the DB direcly, still under conflict protection
func (this *ConcurrentUrl) ReadCommitted(tx uint32, key string) (interface{}, error) {
	if _, err := this.Read(tx, key); err != nil { // For conflict detection
		return nil, err
	}
	return (*this.Store()).Retrive(key)
}

func (this *ConcurrentUrl) Indexer() *ccurltype.Indexer {
	return this.indexer
}

func (this *ConcurrentUrl) Init(store ccurlcommon.DatastoreInterface) {
	this.indexer.Init(store)
	this.invIndexer.Init(store)
	this.reset()
}

func (this *ConcurrentUrl) reset() {
	this.records = this.records[:0]
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
	paths, syspaths, err := this.Platform.Builtin(platform, acct)
	for _, p := range paths {
		path := syspaths[p]
		var v interface{}
		switch path.ID {
		case ccurlcommon.CommutativeMeta: // Path
			v, err = commutative.NewMeta(path.Default.(string))

		case uint8(reflect.Kind(ccurlcommon.NoncommutativeString)): // delta big int
			v = noncommutative.NewString(path.Default.(string))

		case uint8(reflect.Kind(ccurlcommon.CommutativeUint256)): // delta big int
			v = commutative.NewU256(path.Default.(*uint256.Int), big.NewInt(0))

		case uint8(reflect.Kind(ccurlcommon.CommutativeInt64)): // big int pointer
			v = commutative.NewInt64(path.Default.(int64), path.Default.(int64))

		case uint8(reflect.Kind(ccurlcommon.NoncommutativeInt64)): // big int pointer
			v = noncommutative.NewInt64(path.Default.(int64))

		case uint8(reflect.Kind(ccurlcommon.NoncommutativeBytes)): // big int pointer
			v = noncommutative.NewBytes(path.Default.([]byte))
		}

		if !this.indexer.IfExists(p) {
			err = this.indexer.Write(tx, p, v, true) // root path
		}
	}
	return err
}

func (this *ConcurrentUrl) IfExists(path string) bool {
	return this.indexer.IfExists(path)
}

func (this *ConcurrentUrl) TryRead(tx uint32, path string) (interface{}, error) {
	if !this.Permit(tx, path, ccurlcommon.USER_READABLE) {
		return nil, errors.New("Error: No permission to read " + path)
	}

	value, _ := this.indexer.TryRead(tx, path)
	return value, nil
}

func (this *ConcurrentUrl) Read(tx uint32, path string) (interface{}, error) {
	if !this.Permit(tx, path, ccurlcommon.USER_READABLE) {
		return nil, errors.New("Error: No permission to read " + path)
	}
	return this.indexer.Read(tx, path), nil // Read an element
}

func (this *ConcurrentUrl) Write(tx uint32, path string, value interface{}) error {
	return this.write(tx, path, value, false)
}

func (this *ConcurrentUrl) Rewrite(tx uint32, path string, value interface{}) error {
	return this.write(tx, path, value, true)
}

func (this *ConcurrentUrl) write(tx uint32, path string, value interface{}, reset bool) error {
	if !this.Permit(tx, path, ccurlcommon.USER_CREATABLE) {
		return errors.New("Error: No permission to write " + path)
	}

	if value != nil {
		if id := (&ccurltype.Univalue{}).GetTypeID(value); id == uint8(reflect.Invalid) {
			return errors.New("Error: Unknown data type !")
		}
	}
	return this.indexer.Write(tx, path, value, reset)
}

// Read by index
func (this *ConcurrentUrl) ReadAt(tx uint32, path string, idx uint64) (interface{}, error) {
	key, err := this.At(tx, path, idx)
	if err == nil {
		return this.Read(tx, *key)
	}
	return nil, err
}

// Write by index
func (this *ConcurrentUrl) WriteAt(tx uint32, path string, idx uint64, value interface{}) error {
	key, err := this.At(tx, path, idx)
	if err == nil {
		return this.Write(tx, *key, value)
	}
	return err
}

// Get the key by index
func (this *ConcurrentUrl) At(tx uint32, path string, idx uint64) (*string, error) {
	meta, err := this.Read(tx, path) // read the container meta
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(meta).String() != "*commutative.Meta" {
		return nil, errors.New("Error: Wrong path")
	}

	keys := meta.(*commutative.Meta).KeyView()
	if idx >= uint64(len(keys)) {
		return nil, errors.New("Error: Out of range")
	}

	path += keys[idx]
	return &path, nil
}

// It the access is permitted
func (this *ConcurrentUrl) Permit(tx uint32, path string, operation uint8) bool {
	if tx == ccurlcommon.SYSTEM || !this.Platform.OnControlList(path) { // Either by the system or no need to control
		return true
	}

	switch operation {
	case ccurlcommon.USER_READABLE:
		return this.Platform.IsPermitted(path, ccurlcommon.USER_READABLE)

	case ccurlcommon.USER_CREATABLE:
		return (this.Platform.IsPermitted(path, ccurlcommon.USER_CREATABLE) && !this.indexer.IfExists(path)) || // Initialization
			(this.Platform.IsPermitted(path, ccurlcommon.USER_UPDATABLE) && this.indexer.IfExists(path)) // Update

	}
	return false
}

func (this *ConcurrentUrl) Import(transitions []ccurlcommon.UnivalueInterface, args ...interface{}) {
	invtransitions := make([]ccurlcommon.UnivalueInterface, 0, len(transitions))
	for i := 0; i < len(transitions); i++ {
		if transitions[i].GetTransitionType() == ccurlcommon.INVARIATE_TRANSITIONS { // Filter out the invariant transitions first
			invtransitions = append(invtransitions, transitions[i])
			transitions[i] = nil
		}
	}
	common.RemoveIf(&transitions, func(v ccurlcommon.UnivalueInterface) bool { return v == nil })

	common.ParallelExecute(
		func() { this.invIndexer.Import(invtransitions, args...) },
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
	if len(txs) == 0 {
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

	accountMerkle := ccurltype.NewAccountMerkle(this.Platform)
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

// Convert to accesses and transitions
func (this *ConcurrentUrl) convert(records []ccurlcommon.UnivalueInterface, accesseBuf, transitBuf *[]ccurlcommon.UnivalueInterface) {
	*accesseBuf = append(*accesseBuf, make([]ccurlcommon.UnivalueInterface, len(records)-len(*accesseBuf))...)
	*transitBuf = append(*transitBuf, make([]ccurlcommon.UnivalueInterface, len(records)-len(*transitBuf))...)

	numThreads := 1
	if len(records) > 64 {
		numThreads = 4
	}

	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			access, trans := records[i].Export(this.indexer)
			if access != nil {
				(*accesseBuf)[i] = access.(ccurlcommon.UnivalueInterface)
			}
			if trans != nil {
				(*transitBuf)[i] = trans.(ccurlcommon.UnivalueInterface)
			}
		}
	}
	common.ParallelWorker(len(records), numThreads, worker)

	common.RemoveIf(accesseBuf, func(v ccurlcommon.UnivalueInterface) bool { return v == nil })
	common.RemoveIf(transitBuf, func(v ccurlcommon.UnivalueInterface) bool { return v == nil })
}

func (this *ConcurrentUrl) Export(needToSort bool) ([]ccurlcommon.UnivalueInterface, []ccurlcommon.UnivalueInterface) {
	this.indexer.Vectorize(this.indexer.Buffer(), &this.records, false) // Export records
	this.convert(this.records, &this.accesseBuf, &this.transitBuf)      // Convert records to accesses and transitions

	if needToSort { // Sort by path, debug only
		ccurltype.Univalues(this.accesseBuf).Sort()
		ccurltype.Univalues(this.transitBuf).Sort()
	}

	// Not in use yet.
	// objs := make([]interface{}, len(this.accesseBuf))
	// for i := range this.accesseBuf {
	// 	objs[i] = this.accesseBuf[i]
	// }
	// (*this.indexer.Store()).UpdateCacheStats(objs)
	return this.accesseBuf, this.transitBuf
}

type PostProcessFunc func(accesses, transitions []ccurlcommon.UnivalueInterface) ([]ccurlcommon.UnivalueInterface, []ccurlcommon.UnivalueInterface)

func (this *ConcurrentUrl) ExportEncoded(ppf PostProcessFunc) ([][]byte, [][]byte) {
	records, transitions := this.Export(false)
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
