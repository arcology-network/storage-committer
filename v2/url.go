package concurrenturl

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"time"

	common "github.com/HPISTechnologies/common-lib/common"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
	ccurltype "github.com/HPISTechnologies/concurrenturl/v2/type"
	commutative "github.com/HPISTechnologies/concurrenturl/v2/type/commutative"
	noncommutative "github.com/HPISTechnologies/concurrenturl/v2/type/noncommutative"
)

type ConcurrentUrl struct {
	indexer  *ccurltype.Indexer
	Platform *ccurlcommon.Platform
}

func NewConcurrentUrl(store ccurlcommon.DB) *ConcurrentUrl {
	return &ConcurrentUrl{
		indexer:  ccurltype.NewIndexer(store),
		Platform: ccurlcommon.NewPlatform(),
	}
}

// load accounts
func (this *ConcurrentUrl) Preload(tx uint32, platform string, acct string) error {
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
	if !this.Permit(tx, path, ccurlcommon.USER_WRITABLE) {
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
	if tx == ccurlcommon.SYSTEM || !this.Platform.OnList(path) { // Either by the system or no need to control
		return true
	}

	switch operation {
	case ccurlcommon.USER_READABLE:
		return this.Platform.IsPermitted(path, ccurlcommon.USER_READABLE)

	case ccurlcommon.USER_WRITABLE:
		return (this.Platform.IsPermitted(path, ccurlcommon.USER_WRITABLE) && !this.indexer.IfExists(path)) || // Initialization
			(this.Platform.IsPermitted(path, ccurlcommon.USER_UPDATABLE) && this.indexer.IfExists(path)) // Update

	}
	return false
}

func (this *ConcurrentUrl) Commit(transitions []ccurlcommon.UnivalueInterface, txs []uint32) []error {
	this.indexer.Import(transitions)
	paths, states, errs := this.indexer.Commit(txs)
	store := this.indexer.Store()
	(*store).BatchSave(paths, states) // Commit to the state store
	(*store).Clear()
	return errs
}

func (this *ConcurrentUrl) Export(needToSort bool) ([]ccurlcommon.UnivalueInterface, []ccurlcommon.UnivalueInterface) {
	t0 := time.Now()
	records := this.indexer.ToArray(this.indexer.Buffer(), false)
	fmt.Println("this.indexer.ToArray():", time.Since(t0))

	t0 = time.Now()
	recordVec := make([]interface{}, len(records))
	transVec := make([]interface{}, len(records))
	auxTranVec := make([]interface{}, len(records))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			recordVec[i], transVec[i], auxTranVec[i] = records[i].Export(this.indexer)
		}
	}
	common.ParallelWorker(len(records), 4, worker)
	fmt.Println("Export():", time.Since(t0))

	t0 = time.Now()
	// Remove duplicates
	accessDict := make(map[string]ccurlcommon.UnivalueInterface)
	transDict := make(map[string]ccurlcommon.UnivalueInterface)
	auxDict := make(map[string]ccurlcommon.UnivalueInterface)
	for i := range recordVec {
		record := recordVec[i]
		trans := transVec[i]
		auxTran := auxTranVec[i]

		if record != nil {
			accessDict[record.(ccurlcommon.UnivalueInterface).GetPath()] = record.(ccurlcommon.UnivalueInterface)
		}

		if trans != nil {
			transDict[trans.(ccurlcommon.UnivalueInterface).GetPath()] = trans.(ccurlcommon.UnivalueInterface)
		}

		if auxTran != nil {
			auxDict[auxTran.(ccurlcommon.UnivalueInterface).GetPath()] = auxTran.(ccurlcommon.UnivalueInterface)
		}
	}
	fmt.Println("Remove duplicates:", time.Since(t0))

	/* Add the auxiliary transitions to the dictionary*/
	for _, v := range auxDict {
		if _, ok := transDict[v.GetPath()]; !ok {
			access, trans, _ := v.Export(this.indexer)
			accessDict[v.GetPath()] = access.(ccurlcommon.UnivalueInterface)

			if trans != nil {
				transDict[v.GetPath()] = trans.(ccurlcommon.UnivalueInterface)
			}
		}
	}

	accesses := this.indexer.ToArray(&accessDict, needToSort)
	transitions := this.indexer.ToArray(&transDict, needToSort)
	return accesses, transitions
}

func (this *ConcurrentUrl) Indexer() *ccurltype.Indexer {
	return this.indexer
}

func (this *ConcurrentUrl) Print() {
	this.indexer.Print()
}
