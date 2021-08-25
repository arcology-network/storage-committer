package concurrenturl

import (
	"bytes"
	"errors"
	"reflect"
	"sort"

	ccurlcommon "github.com/HPISTechnologies/concurrenturl/common"
	ccurltype "github.com/HPISTechnologies/concurrenturl/type"
	commutative "github.com/HPISTechnologies/concurrenturl/type/commutative"
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
	paths, err := this.Platform.Builtin(platform, acct)
	for _, path := range paths {
		// if this.indexer.Read(tx, path) == nil { // Add the root paths
		if !this.indexer.IfExists(path) {
			meta, _ := commutative.NewMeta(path)
			if meta != nil {
				this.indexer.Write(tx, path, meta) // root path
			}
		}
	}

	return err
}

func (this *ConcurrentUrl) Initialize(platform string, acct string) error {
	pathMeta := this.indexer.Read(ccurlcommon.SYSTEM, platform+acct+"/")
	meta := pathMeta.(*commutative.Meta)
	meta.SetKeys([]string{"code", "nonce", "balance", "defer/", "storage/"})
	this.indexer.Save(meta.Path(), meta)
	return nil
}

func (this *ConcurrentUrl) IfExists(path string) bool {
	return this.indexer.IfExists(path)
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
		return this.Platform.IsPermissible(path, ccurlcommon.USER_READABLE)

	case ccurlcommon.USER_WRITABLE:
		return (this.Platform.IsPermissible(path, ccurlcommon.USER_WRITABLE) && !this.indexer.IfExists(path)) || // Initialization
			(this.Platform.IsPermissible(path, ccurlcommon.USER_UPDATABLE) && this.indexer.IfExists(path)) // Update

	}
	return false
}

func (this *ConcurrentUrl) Commit(transitions []ccurlcommon.UnivalueInterface, txs []uint32) []error {
	this.indexer.Import(transitions)
	return this.indexer.Commit(txs)
}

func (this *ConcurrentUrl) Export() ([]ccurlcommon.UnivalueInterface, []ccurlcommon.UnivalueInterface) {
	records := this.indexer.ToArray()

	accessDict := make(map[string]ccurlcommon.UnivalueInterface)
	transDict := make(map[string]ccurlcommon.UnivalueInterface)
	auxDict := make(map[string]ccurlcommon.UnivalueInterface)
	for i := range records {
		univalue := records[i]
		if univalue.GetTx() == ccurlcommon.SYSTEM && univalue.GetWrites() == 0 {
			continue // system read only less likely to happen
		}

		record, trans, auxTran := univalue.Export(this.indexer)
		if record != nil {
			accessDict[record.(ccurlcommon.UnivalueInterface).GetPath()] = record.(ccurlcommon.UnivalueInterface)
		}

		if trans != nil {
			transDict[trans.(ccurlcommon.UnivalueInterface).GetPath()] = trans.(ccurlcommon.UnivalueInterface)
		}

		if auxTran != nil {
			path := auxTran.(ccurlcommon.UnivalueInterface).GetPath()
			v := auxDict[path]
			if v == nil {
				auxDict[path] = auxTran.(ccurlcommon.UnivalueInterface)
			} else {
				v.Merge(auxTran.(ccurlcommon.UnivalueInterface).GetTx(), auxTran)
				auxDict[path] = v
			}
		}
	}

	/* Add the auxiliary transitions to the dictionary*/
	for _, v := range auxDict {
		if current, ok := transDict[v.GetPath()]; ok {
			current.Merge(v.GetTx(), v)
			transDict[v.GetPath()] = current
			continue
		}
		transDict[v.GetPath()] = v
	}

	/* Finalized the transitions */
	for k, v := range transDict {
		if v.GetValue() != nil {
			v.GetValue().(ccurlcommon.TypeInterface).Transitional(this.indexer.Buffer())
			transDict[k] = v
		}

		if accessDict[k] == nil {
			access, _, _ := v.Export(this.indexer)
			accessDict[k] = access.(ccurlcommon.UnivalueInterface)
		}
	}

	return this.ToArray(&accessDict), this.ToArray(&transDict)
}

/* Map to array */
func (this *ConcurrentUrl) ToArray(dict *map[string]ccurlcommon.UnivalueInterface) []ccurlcommon.UnivalueInterface {
	/* Map to array */
	array := make([]ccurlcommon.UnivalueInterface, 0, len(*dict))
	for _, v := range *dict {
		array = append(array, v)
	}

	sort.SliceStable(array, func(i, j int) bool {
		return bytes.Compare([]byte(array[i].GetPath()[:]), []byte(array[j].GetPath()[:])) < 0
	})
	return array
}

func (this *ConcurrentUrl) Print() {
	this.indexer.Print()
}
