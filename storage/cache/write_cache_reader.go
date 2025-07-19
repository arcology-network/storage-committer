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

// This package includes functions for indexing, retrieving, and manipulating data stored in the cache.
// The cache supports concurrent access and provides methods for reading and writing data at specific paths.
package cache

import (
	"errors"
	"fmt"
	"math"

	common "github.com/arcology-network/common-lib/common"
	deltaset "github.com/arcology-network/common-lib/exp/deltaset"
	stgcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/type/commutative"
	"github.com/arcology-network/storage-committer/type/univalue"
)

// Get the index of a given key under a path.
func (this *WriteCache) IndexOf(tx uint64, path string, key any, T any) (uint64, uint64) {
	if !common.IsPath(path) {
		return math.MaxUint64, 0 //, errors.New("Error: Not a path!!!")
	}

	getter := func(v any) (uint32, uint32, uint32, any) { return 1, 0, 0, v }
	if v, err := this.Do(tx, path, getter, T); err == nil {
		pathInfo := v.(*univalue.Univalue).Value()
		if common.IsType[*commutative.Path](pathInfo) && common.IsType[string](key) {
			// idx, _ := pathInfo.(*commutative.Path).View().IdxOf(key.(string))
			return pathInfo.(*commutative.Path).View().IdxOf(key.(string)), stgcommon.MIN_READ_SIZE
		}
	}
	return math.MaxUint64, stgcommon.MIN_READ_SIZE
}

// KeyAt returns the index of a give key and the the opertion fee under a path.
// If the path does not exist, it returns an error. The second return value is the operation fee.
func (this *WriteCache) KeyAt(tx uint64, path string, index any, T any) (string, uint64) {
	if !common.IsPath(path) {
		return "", stgcommon.MIN_READ_SIZE //, errors.New("Error: Not a path!!!")
	}

	getter := func(v any) (uint32, uint32, uint32, any) { return 1, 0, 0, v }
	if v, err := this.Do(tx, path, getter, T); err == nil {
		pathInfo := v.(*univalue.Univalue).Value()
		if common.IsType[*commutative.Path](pathInfo) && common.IsType[uint64](index) {
			return pathInfo.(*commutative.Path).View().KeyAt(index.(uint64)), stgcommon.MIN_READ_SIZE
		}
	}
	return "", stgcommon.MIN_READ_SIZE
}

// Peek the value under a path. The difference between Peek and Read is that Peek does not have access metadata attached.
func (this *WriteCache) Peek(path string, T any) (any, uint64) {
	v, dataSize := this.PeekRaw(path, T)
	if v == nil {
		return nil, dataSize
	}
	originalV, _, _ := v.(stgcommon.Type).Get()
	return originalV, dataSize
}

// Peek the value under a path. The difference between Peek and Read is that Peek does not have access metadata attached.
func (this *WriteCache) PeekRaw(path string, T any) (any, uint64) {
	_, univ := this.Find(stgcommon.SYSTEM, path, T)
	v := univ.(*univalue.Univalue).Value()
	if v == nil {
		return nil, stgcommon.MIN_READ_SIZE
	}
	return v, v.(stgcommon.Type).MemSize()
}

// // Peek the value under a path. The difference between Peek and Read is that Peek does not have access metadata attached.
// func (this *WriteCache) PeekExist(path string) (bool, uint64) {
// 	_, univ := this.Find(stgcommon.SYSTEM, path, T)
// 	v := univ.(*univalue.Univalue).Value()
// 	return v != nil, stgcommon.MIN_READ_SIZE
// }

// This function looks up the committed value in the DB instead of the cache.
func (this *WriteCache) PeekCommitted(path string, T any) (any, uint64) {
	v, _ := this.backend.Retrive(path, T)
	return v, v.(stgcommon.Type).MemSize()
}

// This function looks up the value and carries out the operation on the value directly.
func (this *WriteCache) Do(tx uint64, path string, doer any, T any) (any, error) {
	univalue, _ := this.RetriveOrCreate(tx, path, T)
	return univalue.Do(tx, path, doer), nil
}

// get the key of the Nth element under a path
func (this *WriteCache) getKeyByIdx(tx uint64, path string, idx uint64) (any, uint64, error) {
	if !common.IsPath(path) {
		return nil, stgcommon.MIN_READ_SIZE, errors.New("Error: Not a path!!!")
	}

	meta, _, dataSize := this.Read(tx, path, new(commutative.Path)) // read the container meta
	length := meta.(*deltaset.DeltaSet[string]).Length()

	if meta != nil {
		subKey := meta.(*deltaset.DeltaSet[string]).KeyAt(idx)
		if len(subKey) > 0 {
			return path + meta.(*deltaset.DeltaSet[string]).KeyAt(idx), dataSize, nil
		}
		return nil, dataSize, errors.New("Error: Exceeded the length of the path !!!")
	}

	return common.IfThen(meta == nil,
		meta,
		common.IfThenDo1st(idx < length,
			func() any { return path + meta.(*deltaset.DeltaSet[string]).KeyAt(idx) },
			nil),
	), dataSize, nil
}

// get the key of the Nth element under a path
func (this *WriteCache) Min(tx uint64, path string, idx uint64) (any, uint64, error) {
	if !common.IsPath(path) {
		return nil, stgcommon.MIN_READ_SIZE, errors.New("Error: Not a path!!!")
	}

	meta, _, dataSize := this.Read(tx, path, new(commutative.Path)) // read the container meta
	if meta != nil {
		if subkey := meta.(*deltaset.DeltaSet[string]).KeyAt(idx); len(subkey) > 0 {
			return path + subkey, dataSize, nil
		}
	}
	return "", dataSize, errors.New("Error: Key not found in the path!!!")
}

// get the key of the Nth element under a path
func (this *WriteCache) Max(tx uint64, path string, idx uint64) (any, uint64, error) {
	if !common.IsPath(path) {
		return nil, stgcommon.MIN_READ_SIZE, errors.New("Error: Not a path!!!")
	}

	meta, _, dataSize := this.Read(tx, path, new(commutative.Path)) // read the container meta
	if meta != nil {
		if subkey := meta.(*deltaset.DeltaSet[string]).KeyAt(idx); len(subkey) > 0 {
			return path + subkey, dataSize, nil
		}
	}
	return "", dataSize, errors.New("Error: Key not found in the path!!!")
}

// Read th Nth element under a path
func (this *WriteCache) ReadAt(tx uint64, path string, idx uint64, T any) (any, uint64, error) {
	if key, dataSize, err := this.getKeyByIdx(tx, path, idx); err == nil && key != nil {
		v, _, readGas := this.Read(tx, key.(string), T)
		return v, dataSize + readGas, nil
	} else {
		return key, dataSize, err
	}
}

// Read th Nth element under a path
func (this *WriteCache) DoAt(tx uint64, path string, idx uint64, do any, T any) (any, uint64, error) {
	if key, dataSize, err := this.getKeyByIdx(tx, path, idx); err == nil && key != nil {
		v, err := this.Do(tx, key.(string), do, T)
		return v, dataSize, err
	} else {
		return key, dataSize, err
	}
}

// Read th Nth element under a path
func (this *WriteCache) PopBack(tx uint64, path string, T any) (any, int64, error) {
	if !common.IsPath(path) {
		return nil, int64(stgcommon.MIN_READ_SIZE), errors.New("Error: Not a path!!!")
	}

	meta, _, _1stReadSize := this.Read(tx, path, T) // read the container meta
	subkey, ok := meta.(*deltaset.DeltaSet[string]).Last()
	if !ok {
		return nil, int64(_1stReadSize), errors.New("Error: The path is either empty or doesn't exist")
	}

	key := path + subkey // Concatenate the path and the subkey
	value, _, _2ndReadSize := this.Read(tx, key, T)
	if value == nil {
		return nil, int64(_1stReadSize + _2ndReadSize), errors.New("Error: Empty container!")
	}

	writeDataSize, err := this.Write(tx, key, nil)
	return value, int64(_1stReadSize+_2ndReadSize) + writeDataSize, err
}

// Remove all the enties in a path, without a single read operation.
// The length will stay the same, but the container will be empty. This is useful for avoiding meta level
// conflicts when the container is appended.
func (this *WriteCache) EraseAll(tx uint64, path string, T any) (any, int64, error) {
	if !common.IsPath(path) {
		return nil, int64(stgcommon.MIN_READ_SIZE), errors.New("Error: Not a path!!!")
	}

	meta, readDataSize := this.Peek(path, T) // read the container meta

	// var accumReadGas uint64
	var accumWriteDataSize int64
	for _, subkey := range meta.(*deltaset.DeltaSet[string]).Elements() {
		key := path + subkey // Concatenate the path and the subkey
		writeData, err := this.Write(tx, key, nil)
		if err != nil {
			fmt.Printf("----------storage-committer/storage/cache/write_cache_reader.go----EraseAll for--key:%v--err:%v \n", key, err)
			// panic(err)
		}
		accumWriteDataSize += writeData
	}
	return nil, int64(readDataSize) + accumWriteDataSize, nil
}

// Read th Nth element under a path
// The way to do this is to use the keys in in the path first and then use the index to get the key.
// Eventually, the key is used to read or write the data. This solution has some issues.
// To get the key by index, the keys in container must be finalized first. If the path
// has any update at this moment, this operation will generate a path read with will conflict with the path write.
func (this *WriteCache) WriteAt(tx uint64, path string, idx uint64, T any) (int64, error) {
	if !common.IsPath(path) {
		return int64(stgcommon.MIN_WRITE_SIZE), errors.New("Error: Not a path!!!")
	}

	if key, readDataSize, err := this.getKeyByIdx(tx, path, idx); err == nil {
		writeDataSize, err := this.Write(tx, key.(string), T)
		return int64(readDataSize) + writeDataSize, err
	} else {
		return int64(readDataSize), err
	}
}
