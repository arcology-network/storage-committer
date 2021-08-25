package common

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/HPISTechnologies/common-lib/common"
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

func (*DataStore) Clear() {}

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
