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
	"reflect"
	"sort"
	"strings"

	common "github.com/arcology-network/common-lib/common"
	mapi "github.com/arcology-network/common-lib/exp/map"
	mempool "github.com/arcology-network/common-lib/exp/mempool"
	slice "github.com/arcology-network/common-lib/exp/slice"
	committercommon "github.com/arcology-network/storage-committer/common"
	platform "github.com/arcology-network/storage-committer/platform"

	importer "github.com/arcology-network/storage-committer/committer/importer"
	"github.com/arcology-network/storage-committer/commutative"
	"github.com/arcology-network/storage-committer/interfaces"
	intf "github.com/arcology-network/storage-committer/interfaces"
	univalue "github.com/arcology-network/storage-committer/univalue"
)

// WriteCache is a read-only data backend used for caching.
type WriteCache struct {
	backend  intf.ReadOnlyDataStore
	kvDict   map[string]*univalue.Univalue // Local KV lookup
	platform intf.Platform
	pool     *mempool.Mempool[*univalue.Univalue]
}

// NewWriteCache creates a new instance of WriteCache; the backend can be another instance of WriteCache,
// resulting in a cascading-like structure.
func NewWriteCache(backend intf.ReadOnlyDataStore, perPage int, numPages int, args ...interface{}) *WriteCache {
	return &WriteCache{
		backend:  backend,
		kvDict:   make(map[string]*univalue.Univalue),
		platform: platform.NewPlatform(),
		pool: mempool.NewMempool(perPage, numPages, func() *univalue.Univalue {
			return new(univalue.Univalue)
		}, (&univalue.Univalue{}).Reset),
	}
}

func (this *WriteCache) SetReadOnlyDataStore(backend intf.ReadOnlyDataStore) *WriteCache {
	this.backend = backend
	return this
}

func (this *WriteCache) ReadOnlyDataStore() intf.ReadOnlyDataStore { return this.backend }
func (this *WriteCache) Cache() *map[string]*univalue.Univalue     { return &this.kvDict }

// func (this *WriteCache) MinSize() int                              { return this.pool.MinSize() }
func (this *WriteCache) NewUnivalue() *univalue.Univalue { return this.pool.New() }

// If the access has been recorded
func (this *WriteCache) GetOrNew(tx uint32, path string, T any) (*univalue.Univalue, bool) {
	unival, inCache := this.kvDict[path]
	if unival == nil { // Not in the kvDict, check the datastore
		var typedv interface{}
		if backend := this.ReadOnlyDataStore(); backend != nil {
			typedv = common.FilterFirst(backend.Retrive(path, T))
		}

		unival = this.NewUnivalue().Init(tx, path, 0, 0, 0, typedv, this)
		this.kvDict[path] = unival // Adding to kvDict
	}
	return unival, inCache // From cache
}

func (this *WriteCache) Read(tx uint32, path string, T any) (interface{}, interface{}, uint64) {
	univalue, _ := this.GetOrNew(tx, path, T)
	return univalue.Get(tx, path, nil), univalue, 0
}

func (this *WriteCache) write(tx uint32, path string, value interface{}) error {
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
	return errors.New("Error: The parent path doesn't exist: " + parentPath)
}

func (this *WriteCache) Write(tx uint32, path string, value interface{}) (int64, error) {
	fee := int64(0) //Fee{}.Writer(path, value, this.writeCache)
	if value == nil || (value != nil && value.(interfaces.Type).TypeID() != uint8(reflect.Invalid)) {
		return fee, this.write(tx, path, value)
	}
	return fee, errors.New("Error: Unknown data type !")
}

// Get data from the DB direcly, still under conflict protection
func (this *WriteCache) ReadCommitted(tx uint32, key string, T any) (interface{}, uint64) {
	if v, _, Fee := this.Read(tx, key, this); v != nil { // For conflict detection
		return v, Fee
	}

	v, _ := this.ReadOnlyDataStore().Retrive(key, T)
	if v == nil {
		return v, 0 //Fee{}.Reader(univalue.NewUnivalue(tx, key, 1, 0, 0, v, nil))
	}
	return v, 0 //Fee{}.Reader(univalue.NewUnivalue(tx, key, 1, 0, 0, v.(interfaces.Type), nil))
}

// Get the raw value directly, skip the access counting at the univalue level
func (this *WriteCache) InCache(path string) (interface{}, bool) {
	univ, ok := this.kvDict[path]
	return univ, ok
}

// Get the raw value directly, put it in an empty univalue without recording the access at the univalue level.
func (this *WriteCache) Find(tx uint32, path string, T any) (interface{}, interface{}) {
	if univ, ok := this.kvDict[path]; ok {
		return univ.Value(), univ
	}

	v, _ := this.ReadOnlyDataStore().Retrive(path, T)
	univ := univalue.NewUnivalue(tx, path, 0, 0, 0, v, nil)
	return univ.Value(), univ
}

// Get the raw value directly.
func (this *WriteCache) Retrive(path string, T any) (interface{}, error) {
	typedv, _ := this.Find(committercommon.SYSTEM, path, T)
	if typedv == nil || typedv.(intf.Type).IsDeltaApplied() {
		return typedv, nil
	}

	rawv, _, _ := typedv.(intf.Type).Get()
	return typedv.(intf.Type).New(rawv, nil, nil, typedv.(intf.Type).Min(), typedv.(intf.Type).Max()), nil // Return in a new univalue
}

func (this *WriteCache) RetriveFromStorage(path string, T any) (interface{}, error) {
	return this.ReadOnlyDataStore().Retrive(path, T)
}

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
	return this.backend.IfExists(path) //this.RetriveShallow(path, nil) != nil
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
	newPathCreations = univalue.Univalues(importer.Sorter(newPathCreations))
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
func (this *WriteCache) Precommit(args ...interface{}) [32]byte {
	this.Insert(args[0].([]*univalue.Univalue))
	return [32]byte{}
}

// Reset the writecache to the initial state for the next round of processing.
func (this *WriteCache) Clear() *WriteCache {
	// if clear(this.buffer); cap(this.buffer) > 3*this.pool.MinSize() {
	// 	this.buffer = make([]*univalue.Univalue, 0, this.pool.MinSize())
	// }
	// this.buffer = this.buffer[:0]
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

func (this *WriteCache) ExportAll(preprocs ...func([]*univalue.Univalue) []*univalue.Univalue) ([]*univalue.Univalue, []*univalue.Univalue) {
	all := this.Export(importer.Sorter)
	// univalue.Univalues(all).Print()

	accesses := univalue.Univalues(slice.Clone(all)).To(importer.ITAccess{})
	transitions := univalue.Univalues(slice.Clone(all)).To(importer.ITTransition{})
	return accesses, transitions
}

func (this *WriteCache) KVs() ([]string, []intf.Type) {
	transitions := univalue.Univalues(slice.Clone(this.Export(importer.Sorter))).To(importer.ITTransition{})

	values := make([]intf.Type, len(transitions))
	keys := slice.ParallelTransform(transitions, 4, func(i int, v *univalue.Univalue) string {
		values[i] = v.Value().(intf.Type)
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
