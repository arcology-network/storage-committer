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

	// performance "github.com/arcology-network/common-lib/mhasher"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type ConcurrentUrl struct {
	importer    *indexer.Importer
	invImporter *indexer.Importer
	writeCache  *indexer.LocalCache
	Platform    *Platform

	ImportFilters []ccurlcommon.FilterTransitionsInterface
	numThreads    int
}

func NewConcurrentUrl(store ccurlcommon.DatastoreInterface, args ...interface{}) *ConcurrentUrl {
	platform := NewPlatform()
	return &ConcurrentUrl{
		Platform: platform,

		importer:      indexer.NewImporter(store, platform),
		invImporter:   indexer.NewImporter(store, platform),
		ImportFilters: []ccurlcommon.FilterTransitionsInterface{&indexer.NonceFilter{}, &indexer.BalanceFilter{}},

		writeCache: indexer.NewLocalCache(store, platform),
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

func (this *ConcurrentUrl) Importer() *indexer.Importer {
	return this.importer
}

func (this *ConcurrentUrl) Init(store ccurlcommon.DatastoreInterface) {
	this.importer.Init(store)
	this.invImporter.Init(store)
}

func (this *ConcurrentUrl) Store() *ccurlcommon.DatastoreInterface {
	return this.importer.Store()
}

func (this *ConcurrentUrl) Clear() {
	(*this.importer.Store()).Clear()

	this.writeCache.Clear()
	this.importer.Clear()
	this.invImporter.Clear()
}

// load accounts
func (this *ConcurrentUrl) CreateAccount(tx uint32, platform string, acct string) error {
	paths, typeids := this.Platform.GetBuiltins(acct)

	var err error
	for i, path := range paths {
		var v interface{}
		switch typeids[i] {
		case commutative.PATH: // Path
			v = commutative.NewPath()

		case uint8(reflect.Kind(noncommutative.STRING)): // delta big int
			v = noncommutative.NewString("")

		case uint8(reflect.Kind(commutative.UINT256)): // delta big int
			v = commutative.NewU256(commutative.U256_MIN, commutative.U256_MAX)

		case uint8(reflect.Kind(commutative.UINT64)):
			v = commutative.NewUint64(0, math.MaxUint64)

		case uint8(reflect.Kind(noncommutative.INT64)):
			v = noncommutative.NewInt64(0)

		case uint8(reflect.Kind(noncommutative.BYTES)):
			v = noncommutative.NewBytes([]byte{})
		}

		if !this.writeCache.IfExists(path) {
			err = this.writeCache.Write(tx, path, v) // root path

			if !this.writeCache.IfExists(path) {
				err = this.writeCache.Write(tx, path, v) // root path
				panic("Failed to create")
			}
		}
	}
	return err
}

func (this *ConcurrentUrl) IfExists(path string) bool {
	return this.writeCache.IfExists(path)
}

func (this *ConcurrentUrl) Peek(path string) (interface{}, error) {
	value, _ := this.writeCache.Peek(path)
	return value, nil
}

func (this *ConcurrentUrl) Read(tx uint32, path string) (interface{}, error) {
	return this.writeCache.Read(tx, path), nil
}

func (this *ConcurrentUrl) Write(tx uint32, path string, value interface{}) error {
	if value != nil {
		if id := (&univalue.Univalue{}).GetTypeID(value); id == uint8(reflect.Invalid) {
			return errors.New("Error: Unknown data type !")
		}
	}
	return this.writeCache.Write(tx, path, value)
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

	keys := meta.([]string)
	return common.IfThenDo1st(idx < uint64(len(keys)), func() interface{} { return path + keys[idx] }, nil), nil
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

	subkeys, err := this.Read(tx, path) // read the container meta
	if subkeys == nil || len(subkeys.([]string)) == 0 || err != nil {
		return nil, common.IfThen(err == nil, errors.New("Error: The path is either empty or doesn't exist"), err)
	}

	key := path + subkeys.([]string)[len(subkeys.([]string))-1]

	value, err := this.Read(tx, key)
	if value == nil || err != nil {
		return nil, errors.New("Error: Empty container!")
	}
	return value, this.Write(tx, key, nil)
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

func (this *ConcurrentUrl) unconditional(transition ccurlcommon.UnivalueInterface) bool {
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
		if this.unconditional(transitions[i]) {
			invTransitions = append(invTransitions, transitions[i])
			transitions[i] = nil
		}
	}
	common.RemoveIf(&transitions, func(v ccurlcommon.UnivalueInterface) bool { return v == nil })

	common.ParallelExecute(
		func() { this.invImporter.Import(invTransitions, args...) },
		func() { this.importer.Import(transitions, args...) })
}

func (this *ConcurrentUrl) KVs() ([]string, []interface{}) {
	keys, values := this.importer.KVs()
	invKeys, invVals := this.invImporter.KVs()

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
		func() { this.invImporter.SortTransitions() },
		func() { this.importer.SortTransitions() })
}

func (this *ConcurrentUrl) Precommit(txs []uint32) []error {
	if txs != nil && len(txs) == 0 { // Commit all the transactions when txs == nil
		return []error{}
	}

	common.ParallelExecute(
		func() { this.invImporter.FinalizeStates() },
		func() {
			this.importer.WhilteList(txs)  // Remove all the transitions generated by the conflicting transactions
			this.importer.FinalizeStates() // Finalize states
		},
	)
	return nil
}

func (this *ConcurrentUrl) WriteToDbBuffer() {
	keys, values := this.KVs()
	(*this.importer.Store()).Precommit(keys, values) // save the transitions to the DB buffer
}

func (this *ConcurrentUrl) SaveToDB() {
	store := this.importer.Store()
	(*store).Commit() // Commit to the state store
	this.Clear()
}

func (this *ConcurrentUrl) Commit(txs []uint32) []error {
	if len(txs) == 0 {
		this.Clear()
		return nil
	}
	errs := this.Precommit(txs)
	this.WriteToDbBuffer()
	this.SaveToDB()
	return errs
}

func (this *ConcurrentUrl) AllInOneCommit(transitions []ccurlcommon.UnivalueInterface, txs []uint32) []error {
	t0 := time.Now()

	accountMerkle := indexer.NewAccountMerkle(this.Platform)
	common.ParallelExecute(
		func() { this.importer.Import(transitions) },
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
	k, v := this.importer.KVs()
	encoded := make([][]byte, 0, len(v))
	for _, value := range v {
		encoded = append(encoded, value.(ccurlcommon.UnivalueInterface).GetEncoded())
	}
	accountMerkle.Build(k, encoded)
	fmt.Println("ComputeMerkle:", time.Since(t0))

	t0 = time.Now()
	this.WriteToDbBuffer()
	fmt.Println("Postcommit :--------------------------------", time.Since(t0))

	t0 = time.Now()
	this.SaveToDB()
	fmt.Println("SaveToDB :--------------------------------", time.Since(t0))

	return []error{}
}

func (this *ConcurrentUrl) Export(preprocessors ...func([]ccurlcommon.UnivalueInterface) []ccurlcommon.UnivalueInterface) []ccurlcommon.UnivalueInterface {
	return this.writeCache.Export(preprocessors...)
}

func (this *ConcurrentUrl) ExportAll(preprocessors ...func([]ccurlcommon.UnivalueInterface) []ccurlcommon.UnivalueInterface) ([]ccurlcommon.UnivalueInterface, []ccurlcommon.UnivalueInterface) {
	return univalue.Univalues(common.Clone(this.Export(ccurlcommon.Sorter))).To(univalue.AccessFilters()...),
		univalue.Univalues(common.Clone(this.Export(ccurlcommon.Sorter))).To(univalue.TransitionFilters()...)
}

func (this *ConcurrentUrl) Print() {
	this.writeCache.Print()
}
