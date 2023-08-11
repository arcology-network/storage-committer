package concurrenturl

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"time"

	"github.com/arcology-network/common-lib/common"
	performance "github.com/arcology-network/common-lib/mhasher"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	interfaces "github.com/arcology-network/concurrenturl/interfaces"
	"github.com/arcology-network/concurrenturl/noncommutative"
	ccurlstorage "github.com/arcology-network/concurrenturl/storage"
	"github.com/arcology-network/concurrenturl/univalue"
)

type ConcurrentUrl struct {
	writeCache  *indexer.WriteCache
	importer    *indexer.Importer
	invImporter *indexer.Importer // transitions that will take effect anyway regardless of execution failures or conflicts
	Platform    *ccurlcommon.Platform

	// ImportFilters []func(unival interfaces.Univalue) interfaces.Univalue
}

func NewConcurrentUrl(store interfaces.Datastore, args ...interface{}) *ConcurrentUrl {
	platform := ccurlcommon.NewPlatform()
	return &ConcurrentUrl{
		writeCache:  indexer.NewWriteCache(store, platform),
		importer:    indexer.NewImporter(store, platform),
		invImporter: indexer.NewImporter(store, platform),
		Platform:    platform, //[]ccurlcommon.FilteredTransitionsInterface{&indexer.NonceFilter{}, &indexer.BalanceFilter{}},
	}
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

	sortedKeys, err := performance.SortStrings(append(keys, invKeys...)) // Keys should be unique
	if err != nil {
		panic(err)
	}
	// sortedKeys := append(keys, invKeys...)
	// sort.Strings(sortedKeys)

	sortedVals := make([]interface{}, len(sortedKeys))
	sorter := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			sortedVals[i] = kvs[sortedKeys[i]]
		}
	}
	common.ParallelWorker(len(sortedKeys), 6, sorter)

	return sortedKeys, sortedVals
}
func (this *ConcurrentUrl) New2(args ...interface{}) *ConcurrentUrl {
	return &ConcurrentUrl{
		writeCache: args[0].(*indexer.WriteCache),
		Platform:   ccurlcommon.NewPlatform(),
	}
}

func (this *ConcurrentUrl) WriteCache() *indexer.WriteCache { return this.writeCache }
func (this *ConcurrentUrl) Importer() *indexer.Importer     { return this.importer }

// Get data from the DB direcly, still under conflict protection
func (this *ConcurrentUrl) ReadCommitted(tx uint32, key string) (interface{}, uint64) {
	if v, Fee := this.Read(tx, key); v != nil { // For conflict detection
		return v, Fee
	}

	v, _ := this.WriteCache().Store().Retrive(key)
	return v, Fee{}.Reader(univalue.NewUnivalue(tx, key, 1, 0, 0, v))
}

func (this *ConcurrentUrl) Init(store interfaces.Datastore) {
	this.importer.Init(store)
	this.invImporter.Init(store)
}

func (this *ConcurrentUrl) Clear() {
	this.importer.Store().Clear()

	this.writeCache.Clear()
	this.importer.Clear()
	this.invImporter.Clear()
}

// load accounts
func (this *ConcurrentUrl) NewAccount(tx uint32, acct string) error {
	paths, typeids := this.Platform.GetBuiltins(acct)

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
			if err := this.writeCache.Write(tx, path, v, true); err != nil { // root path
				return err
			}

			if !this.writeCache.IfExists(path) {
				return this.writeCache.Write(tx, path, v, true) // root path
			}
		}
	}
	return nil
}

func (this *ConcurrentUrl) IfExists(path string) bool {
	return this.writeCache.IfExists(path)
}

func (this *ConcurrentUrl) IndexOf(tx uint32, path string, key interface{}) (uint64, uint64) {
	if !common.IsPath(path) {
		return math.MaxUint64, READ_NONEXIST //, errors.New("Error: Not a path!!!")
	}

	getter := func(v interface{}) (uint32, uint32, uint32, interface{}) { return 1, 0, 0, v }
	if v, err := this.Do(tx, path, getter); err == nil {
		pathInfo := v.(interfaces.Univalue).Value()
		if common.IsType[*commutative.Path](pathInfo) && common.IsType[string](key) {
			return pathInfo.(*commutative.Path).View().IdxOf(key.(string)), 0
		}
	}
	return math.MaxUint64, READ_NONEXIST
}

func (this *ConcurrentUrl) KeyAt(tx uint32, path string, index interface{}) (string, uint64) {
	if !common.IsPath(path) {
		return "", READ_NONEXIST //, errors.New("Error: Not a path!!!")
	}

	getter := func(v interface{}) (uint32, uint32, uint32, interface{}) { return 1, 0, 0, v }
	if v, err := this.Do(tx, path, getter); err == nil {
		pathInfo := v.(interfaces.Univalue).Value()
		if common.IsType[*commutative.Path](pathInfo) && common.IsType[uint64](index) {
			return pathInfo.(*commutative.Path).View().KeyAt(index.(uint64)), 0
		}
	}
	return "", READ_NONEXIST
}

func (this *ConcurrentUrl) Peek(path string) (interface{}, uint64) {
	typedv, univ := this.writeCache.Peek(path)
	return typedv, Fee{}.Reader(univ.(interfaces.Univalue))
}

func (this *ConcurrentUrl) PeekCommitted(path string) (interface{}, uint64) {
	v := this.writeCache.RetriveShallow(path)
	return v, READ_COMMITTED_FROM_DB
}

func (this *ConcurrentUrl) Read(tx uint32, path string) (interface{}, uint64) {
	typedv, univ := this.writeCache.Read(tx, path)
	return typedv, Fee{}.Reader(univ.(interfaces.Univalue))
}

func (this *ConcurrentUrl) Write(tx uint32, path string, value interface{}, persistent bool) (int64, error) {
	fee := Fee{}.Writer(path, value, this.writeCache)
	if value == nil || (value != nil && value.(interfaces.Type).TypeID() != uint8(reflect.Invalid)) {
		return fee, this.writeCache.Write(tx, path, value, persistent)
	}

	return fee, errors.New("Error: Unknown data type !")
}

func (this *ConcurrentUrl) Do(tx uint32, path string, doer interface{}) (interface{}, error) {
	return this.writeCache.Do(tx, path, doer), nil
}

// Read th Nth element under a path
func (this *ConcurrentUrl) at(tx uint32, path string, idx uint64) (interface{}, uint64, error) {
	if !common.IsPath(path) {
		return nil, READ_NONEXIST, errors.New("Error: Not a path!!!")
	}

	meta, readFee := this.Read(tx, path) // read the container meta
	return common.IfThen(meta == nil,
		meta,
		common.IfThenDo1st(idx < uint64(len(meta.([]string))), func() interface{} { return path + meta.([]string)[idx] }, nil),
	), readFee, nil
}

// Read th Nth element under a path
func (this *ConcurrentUrl) ReadAt(tx uint32, path string, idx uint64) (interface{}, uint64, error) {
	if key, Fee, err := this.at(tx, path, idx); err == nil && key != nil {
		v, Fee := this.Read(tx, key.(string))
		return v, Fee, nil
	} else {
		return key, Fee, err
	}
}

// Read th Nth element under a path
func (this *ConcurrentUrl) DoAt(tx uint32, path string, idx uint64, do interface{}) (interface{}, uint64, error) {
	if key, Fee, err := this.at(tx, path, idx); err == nil && key != nil {
		v, err := this.Do(tx, key.(string), do)
		return v, Fee, err
	} else {
		return key, Fee, err
	}
}

// Read th Nth element under a path
func (this *ConcurrentUrl) PopBack(tx uint32, path string, persistent bool) (interface{}, int64, error) {
	if !common.IsPath(path) {
		return nil, int64(READ_NONEXIST), errors.New("Error: Not a path!!!")
	}

	subkeys, Fee := this.Read(tx, path) // read the container meta
	if subkeys == nil || len(subkeys.([]string)) == 0 {
		return nil, int64(Fee), errors.New("Error: The path is either empty or doesn't exist")
	}

	key := path + subkeys.([]string)[len(subkeys.([]string))-1]

	value, Fee := this.Read(tx, key)
	if value == nil {
		return nil, int64(Fee), errors.New("Error: Empty container!")
	}

	writeFee, err := this.Write(tx, key, nil, persistent)
	return value, writeFee, err
}

// Read th Nth element under a path
func (this *ConcurrentUrl) WriteAt(tx uint32, path string, idx uint64, value interface{}, persistent bool) (int64, error) {
	if !common.IsPath(path) {
		return int64(READ_NONEXIST), errors.New("Error: Not a path!!!")
	}

	if key, Fee, err := this.at(tx, path, idx); err == nil {
		return this.Write(tx, key.(string), value, persistent)
	} else {
		return int64(Fee), err
	}
}

func (this *ConcurrentUrl) Import(transitions []interfaces.Univalue, args ...interface{}) *ConcurrentUrl {
	invTransitions := make([]interfaces.Univalue, 0, len(transitions))

	for i := 0; i < len(transitions); i++ {
		if transitions[i].Persistent() {
			invTransitions = append(invTransitions, transitions[i]) //
			transitions[i] = nil
		}
	}
	common.Remove(&transitions, nil)

	common.ParallelExecute(
		func() { this.invImporter.Import(invTransitions, args...) },
		func() { this.importer.Import(transitions, args...) })
	return this
}

func (this *ConcurrentUrl) Snapshot(preTransitions []interfaces.Univalue) interfaces.Datastore {
	// transitions := []interfaces.Univalue(indexer.Univalues(common.Clone(this.Export())).To(indexer.ITCTransition{}))
	// transitions = append(transitions, preTransitions...)

	transientDB := ccurlstorage.NewTransientDB(this.WriteCache().Store()) // Should be the same as Importer().Store()
	preTransitions = common.Remove(&preTransitions, nil)
	snapshotUrl := NewConcurrentUrl(transientDB).Import(preTransitions).Sort()

	ids := indexer.Univalues(preTransitions).UniqueTXs()
	return snapshotUrl.Commit(ids).Importer().Store() // Commit these changes to the a transient DB
}

// Call this as s
func (this *ConcurrentUrl) Sort() *ConcurrentUrl {
	common.ParallelExecute(
		func() { this.invImporter.SortDeltaSequences() },
		func() { this.importer.SortDeltaSequences() })

	return this
}

func (this *ConcurrentUrl) Finalize(txs []uint32) *ConcurrentUrl {
	if txs != nil && len(txs) == 0 { // Commit all the transactions when txs == nil
		return this
	}

	common.ParallelExecute(
		func() { this.invImporter.MergeStateDelta() },
		func() {
			this.importer.WhilteList(txs)   // Remove all the transitions generated by the conflicting transactions
			this.importer.MergeStateDelta() // Finalize states
		},
	)
	return this
}

func (this *ConcurrentUrl) WriteToDbBuffer() {
	keys, values := this.importer.KVs()
	invKeys, invVals := this.invImporter.KVs()

	keys = append(keys, invKeys...)
	values = append(values, invVals...)

	this.importer.Store().Precommit(keys, values) // save the transitions to the DB buffer
}

func (this *ConcurrentUrl) SaveToDB() {
	store := this.importer.Store()
	store.Commit() // Commit to the state store
	this.Clear()
}

func (this *ConcurrentUrl) Commit(txs []uint32) *ConcurrentUrl {
	if txs != nil && len(txs) == 0 {
		this.Clear()
		return this
	}
	this.Finalize(txs)
	this.WriteToDbBuffer()
	this.SaveToDB()
	return this
}

func (this *ConcurrentUrl) AllInOneCommit(transitions []interfaces.Univalue, txs []uint32) []error {
	t0 := time.Now()

	accountMerkle := indexer.NewAccountMerkle(this.Platform)
	common.ParallelExecute(
		func() { this.importer.Import(transitions) },
		func() { accountMerkle.Import(transitions) })

	fmt.Println("indexer.Import + accountMerkle Import :--------------------------------", time.Since(t0))

	t0 = time.Now()
	this.Sort()
	fmt.Println("indexer.Commit :--------------------------------", time.Since(t0))

	t0 = time.Now()
	runtime.GC()
	fmt.Println("GC 0:--------------------------------", time.Since(t0))

	t0 = time.Now()
	this.Finalize(txs)
	fmt.Println("Precommit :--------------------------------", time.Since(t0))

	// Build the merkle tree
	t0 = time.Now()
	k, v := this.importer.KVs()
	encoded := make([][]byte, 0, len(v))
	for _, value := range v {
		encoded = append(encoded, value.(interfaces.Univalue).GetEncoded())
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

func (this *ConcurrentUrl) Export(preprocessors ...func([]interfaces.Univalue) []interfaces.Univalue) []interfaces.Univalue {
	return this.writeCache.Export(preprocessors...)
}

func (this *ConcurrentUrl) ExportAll(preprocessors ...func([]interfaces.Univalue) []interfaces.Univalue) ([]interfaces.Univalue, []interfaces.Univalue) {
	all := this.Export(indexer.Sorter)
	// indexer.Univalues(all).Print()

	accesses := indexer.Univalues(common.Clone(all)).To(indexer.ITCAccess{})
	transitions := indexer.Univalues(common.Clone(all)).To(indexer.ITCTransition{})

	return accesses, transitions
}

func (this *ConcurrentUrl) Print() {
	this.writeCache.Print()
}
