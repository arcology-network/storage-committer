/*
 *   Copyright (c) 2023 Arcology Network
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package cache

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"

	common "github.com/arcology-network/common-lib/common"
	mapi "github.com/arcology-network/common-lib/exp/map"
	mempool "github.com/arcology-network/common-lib/exp/mempool"
	slice "github.com/arcology-network/common-lib/exp/slice"
	committercommon "github.com/arcology-network/storage-committer/common"
	intf "github.com/arcology-network/storage-committer/common"
	stgtype "github.com/arcology-network/storage-committer/common"
	stgeth "github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/type/commutative"
	univalue "github.com/arcology-network/storage-committer/type/univalue"
)

// WriteCache is a read-only data backend used for caching.
type WriteCache struct {
	backend  intf.ReadOnlyStore
	kvDict   map[string]*univalue.Univalue // Local KV lookup
	platform stgeth.Platform
	pool     *mempool.Mempool[*univalue.Univalue]
}

// NewWriteCache creates a new instance of WriteCache; the backend can be another instance of WriteCache,
// resulting in a cascading-like structure.
func NewWriteCache(backend intf.ReadOnlyStore, perPage int, numPages int, args ...interface{}) *WriteCache {
	return &WriteCache{
		backend:  backend,
		kvDict:   make(map[string]*univalue.Univalue),
		platform: *stgeth.NewPlatform(),
		pool: mempool.NewMempool(perPage, numPages, func() *univalue.Univalue {
			return new(univalue.Univalue)
		}, (&univalue.Univalue{}).Reset),
	}
}

func (this *WriteCache) SetReadOnlyBackend(backend intf.ReadOnlyStore) *WriteCache {
	this.backend = backend
	return this
}

func (this *WriteCache) ReadOnlyStore() intf.ReadOnlyStore     { return this.backend }
func (this *WriteCache) Cache() *map[string]*univalue.Univalue { return &this.kvDict }
func (this *WriteCache) Preload([]byte) interface{}            { return nil } // Placeholder
func (this *WriteCache) NewUnivalue() *univalue.Univalue       { return this.pool.New() }

// If the access has been recorded
func (this *WriteCache) GetOrNew(tx uint64, path string, T any) (*univalue.Univalue, bool) {
	unival, inCache := this.kvDict[path]
	if unival == nil { // Not in the kvDict, check the datastore
		unival = this.GetFromStore(tx, path, T)
	}
	return unival, inCache // From cache
}

func (this *WriteCache) write(tx uint64, path string, value interface{}) error {
	parentPath := common.GetParentPath(path)
	if this.IfExists(parentPath) || tx == committercommon.SYSTEM { // The parent path exists or to inject the path directly
		univalue, inCache := this.GetOrNew(tx, path, value) // Get a univalue wrapper
		err := univalue.Set(tx, path, value, inCache, this)

		// Update the parent path meta
		if err == nil {
			if strings.HasSuffix(parentPath, "/container/") || !this.platform.IsSysPath(parentPath) && tx != committercommon.SYSTEM { // Don't keep track of the system children
				parentMeta, inCache := this.GetOrNew(tx, parentPath, new(commutative.Path))
				err = parentMeta.Set(tx, path, univalue.Value(), inCache, this)
			}
		}
		return err
	}
	return errors.New("Error: The parent path " + parentPath + " doesn't exist for " + path)
}

func (this *WriteCache) Read(tx uint64, path string, T any) (interface{}, interface{}, uint64) {
	univalue, _ := this.GetOrNew(tx, path, T)

	// need to check if it is in the memory. If so gas price should be 3 instead.
	gas := uint64(stgtype.GAS_READ)
	if typedv := univalue.Value(); typedv != nil {
		gas += (typedv.(stgtype.Type).MemSize() / 32) * stgtype.GAS_READ
	}

	return univalue.Get(tx, path, nil), univalue, gas
}

func (this *WriteCache) Write(tx uint64, path string, value interface{}) (int64, error) {
	oldSize := uint64(stgtype.GAS_WRITE)
	if v, _ := this.Find(tx, path, value); v != nil {
		oldSize += v.(stgtype.Type).MemSize()
	}

	newSize := uint64(0)
	if value != nil {
		newSize = value.(stgtype.Type).MemSize()
	}

	// Could be negative if the value is deleted or replaced by a value with a smaller size.
	fee := int64(math.Ceil(float64(newSize-oldSize)/32)) * int64(stgtype.GAS_WRITE)
	if value == nil || (value != nil && value.(stgtype.Type).TypeID() != uint8(reflect.Invalid)) {
		return fee, this.write(tx, path, value)
	}
	return fee, errors.New("Error: Unknown data type !")
}

// Read the value from the writecache or the backend. This function is used for
// GetCommittedState() in Eth interface. It is used in gas refund related code.
func (this *WriteCache) ReadCommitted(tx uint64, key string, T any) (interface{}, uint64) {
	// Just to leave a record for conflict detection. This is different from the original Ethereum implementation.
	// In Ethereum, there is no such concept as the multiprocessor，so the committed state can only come from the
	// previous block or the transactions before the current one. But in the multiprocessor, the committed state
	// may also come from the parent thread. So we need to leave a record for the conflict detection in case that
	// threads spawned by multiple parent are trying to access the same path.
	if v := this.GetFromStore(tx, key, this); v != nil { // Check to see if the path exists in the backend.
		return v.Get(tx, key, nil), 0
	}
	return nil, 0
}

// Get the raw value directly, skip the access counting at the univalue level
func (this *WriteCache) InCache(path string) (interface{}, bool) {
	univ, ok := this.kvDict[path]
	return univ, ok
}

// Get the raw value directly, put it in an empty univalue without recording the access at the univalue level.
func (this *WriteCache) Find(tx uint64, path string, T any) (interface{}, interface{}) {
	if univ, ok := this.kvDict[path]; ok {
		return univ.Value(), univ
	}

	v, _ := this.ReadOnlyStore().Retrive(path, T)
	univ := univalue.NewUnivalue(tx, path, 0, 0, 0, v, nil)
	return univ.Value(), univ
}

// The function is used to get the univalue from the backend only.
func (this *WriteCache) GetFromStore(tx uint64, path string, T any) *univalue.Univalue {
	var typedv interface{}
	if backend := this.ReadOnlyStore(); backend != nil {
		typedv, _ = backend.Retrive(path, T)
	}

	unival := this.NewUnivalue().Init(tx, path, 0, 0, 0, typedv, this)
	this.kvDict[path] = unival // Adding to kvDict
	return unival
}

// This function specifically retrieves the value from the backend without any tracking.
func (this *WriteCache) RetriveFromStorage(key string, T any) (interface{}, error) {
	if this.backend == nil {
		return nil, errors.New("Error: The backend is nil")
	}
	return this.backend.RetriveFromStorage(key, T)
}

// Get the raw value directly whichout tracking the accessing record.
// Users need to track the access count themselves.
func (this *WriteCache) Retrive(path string, T any) (interface{}, error) {
	typedv, _ := this.Find(committercommon.SYSTEM, path, T)
	if typedv == nil || typedv.(stgtype.Type).IsDeltaApplied() {
		return typedv, nil
	}

	// Special treatment for the commutative.Path.
	// In general, value types need to be fully cloned as well, so they be
	// manipulated without affecting the original value. But this doesn't apply to the commutative.Path, which
	// has its own change tracking mechanism.
	if common.IsType[*commutative.Path](typedv) {
		return typedv.(*commutative.Path).Clone(), nil
	}

	// Make a Deep copy of the original value.
	rawv, _, _ := typedv.(stgtype.Type).Get()
	return typedv.(stgtype.Type).New(rawv, nil, nil, typedv.(stgtype.Type).Min(), typedv.(stgtype.Type).Max()), nil // Clone the value
}

// Check if the path exists in the writecache or the backend.
func (this *WriteCache) IfExists(path string) bool {
	if committercommon.ETH10_ACCOUNT_PREFIX_LENGTH == len(path) {
		return true
	}

	if v := this.kvDict[path]; v != nil {
		return v.Value() != nil // If value == nil means either it's been deleted or never existed.
	}

	if this.backend == nil {
		return false
	}

	flag := this.backend.IfExists(path) //this.RetriveShallow(path, nil) != nil
	return flag
}

// The function is used to add the transitions to the writecache. It assumes that the transition's
// parent path has been added to the writecache already. Otherwise, it won't succeed.
func (this *WriteCache) set(v *univalue.Univalue) *WriteCache {
	if v == nil {
		return this
	}

	if common.IsPath(*v.GetPath()) && v.Preexist() {
		return this
	}

	(*v).CopyTo(this)
	return this
}

// The function is used to add the transitions to the writecache, which usually comes from
// the child writecaches. It usually happens with the sub processeses are completed.
func (this *WriteCache) Insert(transitions []*univalue.Univalue) *WriteCache {
	if len(transitions) == 0 {
		return this
	}

	// Filter out the path creations transitions as they will be treated differently.
	newPathCreations := slice.MoveIf(&transitions, func(_ int, v *univalue.Univalue) bool {
		return common.IsPath(*v.GetPath()) && !v.Preexist()
	})

	// Not necessary to sort the path creations at the moment,
	// but it is good for the future if multiple level containers are available
	newPathCreations = univalue.Univalues(univalue.Sorter(newPathCreations))
	slice.Foreach(newPathCreations, func(_ int, v **univalue.Univalue) {
		(*v).CopyTo(this) // Write back to the parent writecache
	})

	// Remove the changes to the existing path meta, as they will be updated automatically
	// when inserting or deleting sub elements. This is just simpler and more straightforward
	// than to keep track of the meta changes and merge them back the meta changes.
	transitions = slice.RemoveIf(&transitions, func(_ int, v *univalue.Univalue) bool {
		return common.IsPath(*v.GetPath())
	})

	// Write back to the parent writecache
	slice.Foreach(transitions, func(_ int, v **univalue.Univalue) {
		(*v).CopyTo(this)
	})
	return this
}

// Reset the writecache to the initial state for the next round of processing.
func (this *WriteCache) Clear() *WriteCache {
	this.pool.Reset()
	clear(this.kvDict)
	return this
}

func (this *WriteCache) Equal(other *WriteCache) bool {
	thisBuffer := mapi.Values(this.kvDict)
	sort.SliceStable(thisBuffer, func(i, j int) bool {
		return *thisBuffer[i].GetPath() < *thisBuffer[j].GetPath()
	})

	otherBuffer := mapi.Values(other.kvDict)
	sort.SliceStable(otherBuffer, func(i, j int) bool {
		return *otherBuffer[i].GetPath() < *otherBuffer[j].GetPath()
	})

	cacheFlag := reflect.DeepEqual(thisBuffer, otherBuffer)
	return cacheFlag
}

// Export the content of the writecache to two arrays of univalues.
// One for the accesses and the other for the transitions.
func (this *WriteCache) Export(preprocs ...func([]*univalue.Univalue) []*univalue.Univalue) []*univalue.Univalue {
	buffer := mapi.Values(this.kvDict)
	for _, proc := range preprocs {
		buffer = common.IfThenDo1st(proc != nil, func() []*univalue.Univalue {
			return proc(buffer)
		}, buffer)
	}

	slice.RemoveIf(&buffer, func(_ int, v *univalue.Univalue) bool { return v.Reads() == 0 && v.IsReadOnly() }) // Remove peeks
	return buffer
}

// For the testing purpose, export the content of the writecache to two arrays of univalues and filter.
func (this *WriteCache) ExportAll(preprocs ...func([]*univalue.Univalue) []*univalue.Univalue) ([]*univalue.Univalue, []*univalue.Univalue) {
	all := this.Export()
	// univalue.Univalues(all).Print()

	accesses := univalue.Univalues(slice.Clone(all)).To(univalue.ITAccess{})
	transitions := univalue.Univalues(slice.Clone(all)).To(univalue.ITTransition{})
	return accesses, transitions
}

func (this *WriteCache) KVs() ([]string, []stgtype.Type) {
	transitions := univalue.Univalues(slice.Clone(this.Export(univalue.Sorter))).To(univalue.ITTransition{})

	values := make([]stgtype.Type, len(transitions))
	keys := slice.ParallelTransform(transitions, 4, func(i int, v *univalue.Univalue) string {
		values[i] = v.Value().(stgtype.Type)
		return *v.GetPath()
	})
	return keys, values
}

// This function is used to write the cache to the data source directly to bypass all the intermediate steps,
// including the conflict detection.
func (this *WriteCache) Print() {
	values := mapi.Values(this.kvDict)
	sort.SliceStable(values, func(i, j int) bool {
		return *values[i].GetPath() < *values[j].GetPath()
	})

	for i, elem := range values {
		fmt.Println("Level : ", i)
		elem.Print()
	}
}

// Calculate the checksum of the writecache for integrity check.
func (this *WriteCache) Checksum() [32]byte {
	values := mapi.Values(this.kvDict)
	sort.SliceStable(values, func(i, j int) bool {
		return *values[i].GetPath() < *values[j].GetPath()
	})
	return univalue.Univalues(values).Checksum()
}
