package common

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/arcology-network/common-lib/common"
)

type DataStore struct {
	sharded [256]map[[32]byte]interface{}
}

func NewDataStore() *DataStore {
	var dataStore DataStore
	for i := 0; i < len(dataStore.sharded); i++ {
		dataStore.sharded[i] = make(map[[32]byte]interface{}, 64)
	}
	return &dataStore
}

func (this *DataStore) Save(path string, v interface{}) {
	key := sha256.Sum256([]byte(path))
	if v == nil {
		delete(this.sharded[key[0]], key)
		return
	}
	this.sharded[key[0]][key] = v
}

func (this *DataStore) Retrive(path string) interface{} {
	key := sha256.Sum256([]byte(path))
	if v, ok := this.sharded[key[0]][key]; ok {
		return v
	}
	return nil
}

func (this *DataStore) BatchSave(paths []string, states []interface{}) {
	keys := make([][32]byte, len(paths))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			keys[i] = sha256.Sum256([]byte(paths[i]))
		}
	}
	common.ParallelWorker(len(paths), 4, worker)

	for i, key := range keys {
		this.sharded[key[0]][key] = states[i]
	}
}

func (this *DataStore) Print() {
	keys := [][32]byte{}
	for _, shard := range this.sharded {
		for k, v := range shard {
			if v != nil {
				keys = append(keys, k)
			}
		}
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return bytes.Compare(keys[i][:], keys[j][:]) < 0
	})

	for _, key := range keys {
		fmt.Println("Key: ", key)
		this.sharded[key[0]][key].(TypeInterface).Print()
	}
}

// func (this *DataStore) BatchRetrive(keys []string) []interface{} {
// 	categorized := make([][]uint32, len(this.sharded))
// 	for i := 0; i < len(categorized); i++ {
// 		categorized[keys[i][0]] = append(categorized[keys[i][0]], uint32(i))
// 	}

// 	values := make([][]interface{}, len(this.sharded))
// 	for i := 0; i < len(categorized); i++ {
// 		for j := 0; j < len(categorized[i]); j++ {
// 			idx := categorized[i][j]
// 			values[i] = append(values[i], this.sharded[i][keys[idx]])
// 		}
// 	}

// 	valueArr := make([]interface{}, 0, len(keys))
// 	for i := 0; i < len(values); i++ {
// 		valueArr = append(valueArr, values[i])
// 	}
// 	return valueArr
// }
