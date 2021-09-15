package concurrenturl

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"sort"
	"time"

	"github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	ccurltype "github.com/arcology-network/concurrenturl/v2/type"
	commutative "github.com/arcology-network/concurrenturl/v2/type/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/v2/type/noncommutative"
)

type ConcurrentUrl struct {
	indexer  *ccurltype.Indexer
	Platform *ccurlcommon.Platform
}

func NewConcurrentUrl(store ccurlcommon.DB, args ...interface{}) *ConcurrentUrl {
	platform := ccurlcommon.NewPlatform()
	return &ConcurrentUrl{
		indexer:  ccurltype.NewIndexer(store, platform),
		Platform: platform,
	}
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

		case uint8(reflect.Kind(ccurlcommon.CommutativeBalance)): // delta big int
			v = commutative.NewBalance(path.Default.([]*big.Int)[0], path.Default.([]*big.Int)[1])

		case uint8(reflect.Kind(ccurlcommon.CommutativeInt64)): // big int pointer
			v = commutative.NewInt64(path.Default.(int64), path.Default.(int64))

		case uint8(reflect.Kind(ccurlcommon.NoncommutativeInt64)): // big int pointer
			v = noncommutative.NewInt64(path.Default.(int64))

		case uint8(reflect.Kind(ccurlcommon.NoncommutativeBytes)): // big int pointer
			v = noncommutative.NewBytes(path.Default.([]byte))
		}

		if !this.indexer.IfExists(p) {
			err = this.indexer.Write(tx, p, v) // root path
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
	return this.indexer.TryRead(tx, path), nil // Read an element
}

func (this *ConcurrentUrl) Read(tx uint32, path string) (interface{}, error) {
	if !this.Permit(tx, path, ccurlcommon.USER_READABLE) {
		return nil, errors.New("Error: No permission to read " + path)
	}
	return this.indexer.Read(tx, path), nil // Read an element
}

func (this *ConcurrentUrl) Write(tx uint32, path string, value interface{}) error {
	if !this.Permit(tx, path, ccurlcommon.USER_CREATABLE) {
		return errors.New("Error: No permission to write " + path)
	}

	if value != nil {
		if id := (&ccurltype.Univalue{}).GetTypeID(value); id == uint8(reflect.Invalid) {
			return errors.New("Error: Unknown data type !")
		}
	}
	return this.indexer.Write(tx, path, value)
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

func (this *ConcurrentUrl) Import(transitions []ccurlcommon.UnivalueInterface) {
	this.indexer.Import(transitions)
}

func (this *ConcurrentUrl) Commit(txs []uint32) []error {
	paths, states, errs := this.indexer.Commit(txs)

	store := this.indexer.Store()
	(*store).BatchSave(paths, states) // Commit to the state store

	(*store).Clear()
	this.Indexer().Clear()

	return errs
}

func (this *ConcurrentUrl) AllInOneCommit(transitions []ccurlcommon.UnivalueInterface, txs []uint32) []error {
	t0 := time.Now()
	this.indexer.Import(transitions)
	fmt.Println("indexer Import :--------------------------------", time.Since(t0))

	/*Account Merkle Tree */
	t0 = time.Now()
	accountMerkle := ccurltype.NewAccountMerkle(this.Platform)
	accountMerkle.Import(transitions)
	fmt.Println("accountMerkle Import :--------------------------------", time.Since(t0))

	t0 = time.Now()
	paths, states, errs := this.indexer.Commit(txs)
	fmt.Println("indexer.Commit :--------------------------------", time.Since(t0))

	t0 = time.Now()
	runtime.GC()
	fmt.Println("GC 0:--------------------------------", time.Since(t0))

	// Build the merkle tree
	t0 = time.Now()
	accountMerkle.Build(8, this.Indexer().ByPathFirst())
	fmt.Println("ComputeMerkle:", time.Since(t0))

	t0 = time.Now()
	store := this.indexer.Store()
	(*store).BatchSave(paths, states) // Commit to the state store
	fmt.Println("BatchSave :--------------------------------", time.Since(t0))

	t0 = time.Now()
	(*store).Clear()
	this.Indexer().Clear()
	fmt.Println("Clear :", time.Since(t0))
	return errs
}

func (this *ConcurrentUrl) singleThredExport(records []ccurlcommon.UnivalueInterface, recordVec, transVec []interface{}) {
	for i := 0; i < len(records); i++ {
		recordVec[i], transVec[i] = records[i].Export(this.indexer)
	}
}

func (this *ConcurrentUrl) multiThredExport(records []ccurlcommon.UnivalueInterface, recordVec, transVec []interface{}) {
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			recordVec[i], transVec[i] = records[i].Export(this.indexer)
		}
	}
	common.ParallelWorker(len(records), 4, worker)
}

func (this *ConcurrentUrl) Export(needToSort bool) ([]ccurlcommon.UnivalueInterface, []ccurlcommon.UnivalueInterface) {
	records := this.indexer.ToArray(this.indexer.Buffer(), false)
	recordVec := make([]interface{}, len(records))
	transVec := make([]interface{}, len(records))
	if len(records) < 64 {
		this.singleThredExport(records, recordVec, transVec)
	} else {
		this.multiThredExport(records, recordVec, transVec)
	}

	accesses := make([]ccurlcommon.UnivalueInterface, 0, len(records))
	transitions := make([]ccurlcommon.UnivalueInterface, 0, len(records))
	for i := 0; i < len(records); i++ {
		if recordVec[i] != nil {
			accesses = append(accesses, recordVec[i].(ccurlcommon.UnivalueInterface))
		}

		if transVec[i] != nil {
			transitions = append(transitions, transVec[i].(ccurlcommon.UnivalueInterface))
		}
	}

	if needToSort { // Sort by path, debug only
		sort.SliceStable(accesses, func(i, j int) bool {
			return bytes.Compare([]byte(accesses[i].GetPath()[:]), []byte(accesses[j].GetPath()[:])) < 0
		})

		sort.SliceStable(transitions, func(i, j int) bool {
			return bytes.Compare([]byte(transitions[i].GetPath()[:]), []byte(transitions[j].GetPath()[:])) < 0
		})
	}

	return accesses, transitions
}

func (this *ConcurrentUrl) Indexer() *ccurltype.Indexer {
	return this.indexer
}

func (this *ConcurrentUrl) Print() {
	this.indexer.Print()
}
