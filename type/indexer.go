package urltype

import (
	"errors"
	"fmt"

	ccurlcommon "github.com/arcology-network/concurrenturl/common"
)

type Indexer struct {
	store  ccurlcommon.DB
	buffer map[string]ccurlcommon.UnivalueInterface

	byTx         map[uint32]*map[string]ccurlcommon.UnivalueInterface
	byPath       map[string][]ccurlcommon.UnivalueInterface
	intialStates map[string]ccurlcommon.UnivalueInterface
	importBuffer map[string]ccurlcommon.UnivalueInterface
}

func NewIndexer(store ccurlcommon.DB) *Indexer {
	var indexer Indexer
	indexer.store = store
	indexer.buffer = make(map[string]ccurlcommon.UnivalueInterface, 1024)

	indexer.byTx = make(map[uint32]*map[string]ccurlcommon.UnivalueInterface)
	indexer.byPath = make(map[string][]ccurlcommon.UnivalueInterface)
	indexer.intialStates = make(map[string]ccurlcommon.UnivalueInterface)
	indexer.importBuffer = make(map[string]ccurlcommon.UnivalueInterface)
	return &indexer
}

func (this *Indexer) NewValue(tx uint32, path string, value interface{}) ccurlcommon.UnivalueInterface {
	univalue := NewUnivalue(tx, path, 0, 0, value, this)
	univalue.SetPreexist(this)
	univalue.AddOrDelete = univalue.IsAddOrDelete()
	univalue.Composite = univalue.IsComposite()
	return univalue
}

func (this *Indexer) Store() *ccurlcommon.DB                            { return &this.store }
func (this *Indexer) Buffer() *map[string]ccurlcommon.UnivalueInterface { return &this.buffer }

func (this *Indexer) IfExists(path string) bool {
	return this.buffer[path] != nil || this.Retrive(path) != nil
}

// If the access has been recorded
func (this *Indexer) CheckHistory(tx uint32, path string, ifSave bool) ccurlcommon.UnivalueInterface {
	univalue := this.buffer[path]
	if univalue == nil { // Not in the buffer, check the datastore
		univalue = this.NewValue(tx, path, this.Retrive(path))
		if ifSave {
			this.buffer[path] = univalue
		}
	}
	return univalue
}

func (this *Indexer) Read(tx uint32, path string) interface{} {
	univalue := this.CheckHistory(tx, path, true)
	return univalue.Get(tx, path, this.Buffer())
}

func (this *Indexer) Write(tx uint32, path string, value interface{}) error {
	parentPath := ccurlcommon.GetParentPath(path)
	if this.IfExists(parentPath) || tx == ccurlcommon.SYSTEM { // The parent path exists or inject paths directly
		univalue := this.CheckHistory(tx, path, true)
		return univalue.Set(tx, path, value, this)
	}
	return errors.New("Error: The parent path doesn't exist")
}

func (this *Indexer) Insert(path string, value interface{}) {
	this.buffer[path] = value.(ccurlcommon.UnivalueInterface)
}

func (this *Indexer) Retrive(key string) interface{} {
	if v := this.store.Retrive(key); v != nil {
		return v.(ccurlcommon.TypeInterface).Deepcopy()
	}
	return nil
}

func (this *Indexer) Save(key string, v interface{}) {
	this.store.Save(key, v)
}

// All transitions from one traxiton
func (this *Indexer) Import(txTrans []ccurlcommon.UnivalueInterface) {
	for _, v := range txTrans {
		this.addToBuffers(v)
	}
}

func (this *Indexer) addToBuffers(v ccurlcommon.UnivalueInterface) {
	if value := this.Retrive(v.GetPath()); value != nil {
		state := this.CheckHistory(ccurlcommon.SYSTEM, v.GetPath(), false)
		this.intialStates[v.GetPath()] = state
	}

	if this.byTx[v.GetTx()] == nil {
		txMap := make(map[string]ccurlcommon.UnivalueInterface)
		this.byTx[v.GetTx()] = &txMap
	}

	(*this.byTx[v.GetTx()])[v.GetPath()] = v
	this.byPath[v.GetPath()] = append(this.byPath[v.GetPath()], v)
}

func (this *Indexer) Commit(whitelist []uint32) []error {
	errs := []error{}
	for _, txID := range whitelist {
		if this.byTx[txID] == nil {
			errs = append(errs, errors.New("Unknow Transaction ID: "+fmt.Sprint(txID)))
			continue
		}

		for k := range *this.byTx[txID] {
			(*this.byTx[txID])[k] = nil
		}
	}

	// Apply the transitions
	for k, transitions := range this.byPath {
		updated := false
		for _, v := range transitions {
			if v == nil {
				continue
			}
			updated = true // The initial state will be updated

			initV := this.intialStates[k] // Get the initial value of the variable
			if initV == nil {
				this.intialStates[k] = v // A new variable
				continue
			}
			initV.Merge(ccurlcommon.SYSTEM, v)
		}

		this.intialStates[k].Finalize()
		if !updated {
			this.intialStates[transitions[0].GetPath()] = nil // No updates, no need to write back to the datastore
		}
	}

	// Strip access info
	paths := make([]string, 0, len(this.intialStates))
	states := make([]interface{}, 0, len(this.intialStates))
	for k, intial := range this.intialStates {
		if intial == nil {
			continue
		}

		paths = append(paths, k)

		v := intial.GetValue()
		if v != nil {
			v.(ccurlcommon.TypeInterface).Purge()
		}
		states = append(states, v)
	}

	this.store.BatchSave(paths, states) // Commit to the state store
	this.clear()
	return errs
}

// Clear all
func (this *Indexer) clear() {
	this.buffer = make(map[string]ccurlcommon.UnivalueInterface)
	this.byTx = make(map[uint32]*map[string]ccurlcommon.UnivalueInterface)
	this.byPath = make(map[string][]ccurlcommon.UnivalueInterface)
	this.intialStates = make(map[string]ccurlcommon.UnivalueInterface)
	this.importBuffer = make(map[string]ccurlcommon.UnivalueInterface)
}
