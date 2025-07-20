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

// write_cache.go provides the implementation of WriteCache, a read-only data backend
// designed for caching key-value pairs in the Arcology Network storage committer module.
// It supports efficient retrieval, insertion, and management of cached data, including
// wildcard deletions, memory pooling, and integration with a backend store. The WriteCache
// is optimized for use in concurrent and multi-processor environments.
//
// Note: The WriteCache itself is read-only; all updates are performed by the committer.
//

package cache

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/associative"
	mapi "github.com/arcology-network/common-lib/exp/map"
	mempool "github.com/arcology-network/common-lib/exp/mempool"
	slice "github.com/arcology-network/common-lib/exp/slice"
	stgcommon "github.com/arcology-network/storage-committer/common"
	stgeth "github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/type/commutative"
	univalue "github.com/arcology-network/storage-committer/type/univalue"
)

// WriteCache is a read-only data backend used for caching.
type WriteCache struct {
	backend     stgcommon.ReadOnlyStore
	kvDict      map[string]*univalue.Univalue       // Local KV lookup
	wildcardDel []*associative.Pair[uint64, string] // Paths delete by wildcard
	platform    stgeth.Platform
	pool        *mempool.Mempool[*univalue.Univalue]
}

// NewWriteCache creates a new instance of WriteCache; the backend can be another instance of WriteCache,
// resulting in a cascading-like structure.
func NewWriteCache(backend stgcommon.ReadOnlyStore, perPage int, numPages int, args ...any) *WriteCache {
	return &WriteCache{
		backend:     backend,
		kvDict:      make(map[string]*univalue.Univalue),
		wildcardDel: make([]*associative.Pair[uint64, string], 0),
		platform:    *stgeth.NewPlatform(),
		pool: mempool.NewMempool(perPage, numPages, func() *univalue.Univalue {
			return new(univalue.Univalue)
		}, (&univalue.Univalue{}).Reset),
	}
}

func (this *WriteCache) SetReadOnlyBackend(backend stgcommon.ReadOnlyStore) *WriteCache {
	this.backend = backend
	return this
}

func (this *WriteCache) AddToDict(v *univalue.Univalue)         { this.kvDict[*v.GetPath()] = v }
func (this *WriteCache) ReadOnlyStore() stgcommon.ReadOnlyStore { return this.backend }
func (this *WriteCache) Cache() *map[string]*univalue.Univalue  { return &this.kvDict }
func (this *WriteCache) Preload([]byte) any                     { return nil } // Placeholder
func (this *WriteCache) NewUnivalue() *univalue.Univalue        { return this.pool.New() }

func (this *WriteCache) FilterWildcards(tx uint64, path string, newVal any, args ...any) (bool, uint64) {
	if flag, cleanPath := univalue.IsWildcard(path); flag { // Check if the path is a wildcard path
		this.wildcardDel = append(this.wildcardDel, &associative.Pair[uint64, string]{First: tx, Second: path})
		pathMeta, _, _ := this.Find(tx, cleanPath, true, newVal, nil) // Read the clean path to ensure it exists in the cache
		return true, pathMeta.(*commutative.Path).TotalSize
	}
	return false, 0 // Return true to indicate that the path is valid
}

// Remove the matched entries ALREADY in the write cache as deleted.
func (this *WriteCache) removeDuplicates(path string) {
	wildcard := common.GetParentPath(path)
	for k, v := range this.kvDict {
		if bytes.Equal([]byte(k[:len(wildcard)]), []byte(wildcard)) {
			if v.Value() != nil {
				v.SetValue(nil)
				v.IncrementWrites(1)
				continue
			}
			delete(this.kvDict, k) // Remove the entry from the cache because it is already covered by the wildcard.
		}
	}
}

// PreloadMatched preloads the paths that match the wildcard delete path that are about to be deleted by the
// the current write operation.
// func (this *WriteCache) PreloadMatched(path string, T any) []*univalue.Univalue {
// 	preloaded := make([]*univalue.Univalue, 0)
// 	for _, wildcardPath := range this.wildcardDel {
// 		if len(path) < len(wildcardPath.Second) {
// 			continue
// 		}

// 		if bytes.Equal([]byte(path[:len(wildcardPath.Second)]), []byte(wildcardPath.Second)) {
// 			_, univ, _ := this.Find(wildcardPath.First, path, T, this.AddToDict) // Preload the path
// 			univ.SetValue(nil)                                                   // To indicate t the path has been deleted by the wildcard
// 			univ.IncrementWrites(1)
// 			univ.SetLocal(true) // Mark as local, so only the wildcard will be exported.
// 			preloaded = append(preloaded, univ)
// 		}
// 	}
// 	return preloaded
// }

// Get the raw value directly, put it in an empty univalue without recording
// the access at the univalue level. Won't update the kvDict.
func (this *WriteCache) Find(tx uint64, path string, isRead bool, T any, do func(*univalue.Univalue)) (any, *univalue.Univalue, bool) {
	for _, wildcardPath := range this.wildcardDel {
		if len(path) < len(wildcardPath.Second) {
			continue
		}

		parentPath := common.GetParentPath(wildcardPath.Second) // Ensure the path is a valid parent path
		if parentPath == path[:len(parentPath)] {
		}
	}

	if univ, ok := this.kvDict[path]; ok {
		return univ.Value(), univ, true // From cache
	}

	univ := this.LoadFromCommitted(tx, path, T)
	if do != nil {
		do(univ) // Call the callback function if provided
	}
	return univ.Value(), univ, false
}

func (this *WriteCache) Write(tx uint64, path string, newVal any, args ...any) (int64, error) {
	if newVal != nil && newVal.(stgcommon.Type).TypeID() == uint8(reflect.Invalid) { // Neither a valid replacement nor a delete operation.
		return 0, errors.New("Error: Unknown data type !")
	}

	if matched, size := this.FilterWildcards(tx, path, newVal, args...); matched {
		this.removeDuplicates(path) // Mark all the matched entries in the write cache as deleted
		return int64(size), nil     // If the path is a wildcard, return the size difference
	}
	// this.PreloadMatched(path, newVal) // Preload the wildcard paths to the cache, if any.

	univ, err := this.write(tx, path, newVal)
	sizeDif := this.DiffSize(tx, path, newVal) // Update the size difference
	if len(args) > 0 && args[0] != nil {
		args[0].(func(*univalue.Univalue, int64))(univ, sizeDif) // Call the callback function if provided
	}
	return sizeDif, err
}

func (this *WriteCache) write(tx uint64, path string, value any) (*univalue.Univalue, error) {
	parentPath := common.GetParentPath(path)
	var univ *univalue.Univalue
	if this.IfExists(parentPath) || tx == stgcommon.SYSTEM { // The parent path exists or to inject the path directly
		_, univ, inCache := this.Find(tx, path, false, value, this.AddToDict) // Get a univalue wrapper
		err := univ.Set(tx, path, value, inCache, this)                       // set the new value

		// Update the parent path meta
		if err == nil {
			if strings.HasSuffix(parentPath, "/container/") || !this.platform.IsSysPath(parentPath) && tx != stgcommon.SYSTEM { // Don't keep track of the system children
				_, parentMeta, inCache := this.Find(tx, parentPath, false, new(commutative.Path), this.AddToDict)
				err = parentMeta.Set(tx, path, univ.Value(), inCache, this)
			}
		}
		return univ, err
	}
	return univ, errors.New("Error: The parent path " + parentPath + " doesn't exist for " + path)
}

// WildcardsToUnivalue converts wildcard paths to Univalue for exporting.
func (this *WriteCache) WildcardsToUnivalue() []*univalue.Univalue {
	univs := make([]*univalue.Univalue, 0)
	for _, wildcardPath := range this.wildcardDel {
		newV := univalue.NewUnivalue(wildcardPath.First, wildcardPath.Second, 0, 1, 0, nil, nil)
		newV.SetPreexist(true) // Mark as pre-existing, so it pass through the filter.
		univs = append(univs, newV)
	}
	return univs
}

// Get the raw value directly WITHOUT tracking the accessing record.
// Users need to count access themselves.
func (this *WriteCache) Retrive(path string, T any) (any, error) {
	typedv, _, _ := this.Find(stgcommon.SYSTEM, path, true, T, nil)
	if typedv == nil || typedv.(stgcommon.Type).IsDeltaApplied() {
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
	rawv, _, _ := typedv.(stgcommon.Type).Get()
	min, max := typedv.(stgcommon.Type).Limits()
	return typedv.(stgcommon.Type).New(rawv, nil, nil, min, max), nil // Clone the value
}

// The load the data from the backend. Since the state is already committed, it is read-only.
// No need to add it to the kvDict or keep track of the access.
func (this *WriteCache) LoadFromCommitted(tx uint64, path string, T any) *univalue.Univalue {
	var typedv any
	if backend := this.ReadOnlyStore(); backend != nil {
		typedv, _ = backend.Retrive(path, T)
	}
	return this.NewUnivalue().Init(tx, path, 0, 0, 0, typedv, this)
}

// This function specifically retrieves the value from the backend without any tracking.
func (this *WriteCache) RetriveFromStorage(key string, T any) (any, error) {
	if this.backend == nil {
		return nil, errors.New("Error: The backend is nil")
	}
	return this.backend.RetriveFromStorage(key, T)
}

func (this *WriteCache) Read(tx uint64, path string, T any) (any, any, uint64) {
	_, univalue, _ := this.Find(tx, path, true, T, this.AddToDict) // Get the univalue wrapper

	// need to check if it is in the memory. If so gas price should be 3 instead.
	dataSize := stgcommon.MIN_READ_SIZE
	if typedv := univalue.Value(); typedv != nil {
		dataSize = typedv.(stgcommon.Type).MemSize()
	}

	return univalue.Get(tx, path, nil), univalue, dataSize
}

func (this *WriteCache) DiffSize(tx uint64, path string, newVal any) int64 {
	oldSize := int64(0)
	if oldVal, _, _ := this.Find(tx, path, true, newVal, nil); oldVal != nil {
		oldSize += int64(oldVal.(stgcommon.Type).MemSize())
	}

	newSize := int64(0)
	if newVal != nil {
		newSize = int64(newVal.(stgcommon.Type).MemSize())
	}

	return newSize - oldSize
}

// Get the raw value directly, skip the access counting at the univalue level
func (this *WriteCache) InCache(path string) (any, bool) {
	univ, ok := this.kvDict[path]
	return univ, ok
}

// Check if the path exists in the writecache or the backend.
// No access count is recorded. Only for internal use. Not exposed to the public API.
func (this *WriteCache) IfExists(path string) bool {
	// Any path shorter than the ETH10_ACCOUNT_PREFIX is a system path.
	if stgcommon.ETH10_ACCOUNT_PREFIX_LENGTH >= len(path) {
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
	slice.RemoveIf(&buffer, func(_ int, v *univalue.Univalue) bool {
		return v.PathLookupOnly() || v.IsLocal() // Remove peeks and local values
	})

	// univalue.Univalues(buffer).PrintUnsorted() // For debugging purpose
	buffer = append(buffer, this.WildcardsToUnivalue()...)
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

func (this *WriteCache) KVs() ([]string, []stgcommon.Type) {
	transitions := univalue.Univalues(slice.Clone(this.Export(univalue.Sorter))).To(univalue.ITTransition{})

	values := make([]stgcommon.Type, len(transitions))
	keys := slice.ParallelTransform(transitions, 4, func(i int, v *univalue.Univalue) string {
		values[i] = v.Value().(stgcommon.Type)
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

// Read the value from the backend. This function is used for
// GetCommittedState() in Eth interface for gas refund related code.
func (this *WriteCache) ReadCommitted(tx uint64, key string, T any) (any, uint64) {
	// Just to leave a record for conflict detection. This is different from the original Ethereum implementation.
	// In Ethereum, there is no such concept as the multiprocessor，so the committed state can only come from the
	// previous block or the transactions before the current one. But in the multiprocessor, the committed state
	// may also come from the parent thread. So we need to leave a record for the conflict detection in case that
	// threads spawned by multiple parent are trying to access the same path.
	if v := this.LoadFromCommitted(tx, key, this); v != nil { // Check to see if the path exists in the backend.
		return v.Get(tx, key, nil), 0
	}
	return nil, 0
}
