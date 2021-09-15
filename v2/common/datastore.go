package common

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/arcology-network/common-lib/common"
	orderedmap "github.com/elliotchance/orderedmap"
)

type DataStore struct {
	numShards uint32
	sharded   []map[string]interface{}
}

func NewDataStore(args ...interface{}) *DataStore {
	dataStore := DataStore{
		numShards: 6,
	}

	dataStore.sharded = make([]map[string]interface{}, dataStore.numShards)
	for i := 0; i < int(dataStore.numShards); i++ {
		dataStore.sharded[i] = make(map[string]interface{}, 64)
	}

	return &dataStore
}

func (this *DataStore) Save(key string, v interface{}) {
	if v == nil {
		delete(this.sharded[key[0]], key)
		return
	}

	var total uint32 = 0
	for j := 0; j < len(key); j++ {
		total += uint32(key[j])
	}

	this.sharded[total%this.numShards][key] = v
}

func (this *DataStore) Retrive(key string) interface{} {
	var total uint32 = 0
	for j := 0; j < len(key); j++ {
		total += uint32(key[j])
	}

	if v, ok := this.sharded[total%this.numShards][key]; ok {
		return v
	}
	return nil
}

func (*DataStore) Clear() {}

func (this *DataStore) BatchSave(keys []string, state interface{}) {
	dict := state.(*map[string]*orderedmap.OrderedMap)
	ids := make([]uint32, len(keys))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			if _, ok := (*dict)[keys[i]]; ok { // Found the element in the dict
				key := keys[i]
				var total uint32 = 0
				for j := 0; j < len(key); j++ {
					total += uint32(key[j])
				}
				ids[i] = total % this.numShards
			} else {
				ids[i] = math.MaxInt32
			}
		}
	}
	common.ParallelWorker(len(keys), 8, worker)

	var wg sync.WaitGroup
	for threadID := 0; threadID < int(this.numShards); threadID++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			for i := 0; i < len(keys); i++ {
				if ids[i] == uint32(threadID) {
					this.sharded[threadID][keys[i]] = (*dict)[keys[i]].Front().Value.(UnivalueInterface).Value()
				}
			}
		}(threadID)
	}
	wg.Wait()
}

func (this *DataStore) _BatchSave(keys []string, transitions []interface{}) {
	t0 := time.Now()
	ids := make([]uint32, len(transitions))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			if transitions[i] != nil {
				key := keys[i]
				var total uint32 = 0
				for j := 0; j < len(key); j++ {
					total += uint32(key[j])
				}
				ids[i] = total % this.numShards
			}
		}
	}
	common.ParallelWorker(len(transitions), 8, worker)
	fmt.Println("total % 4 "+fmt.Sprint(100000*9), time.Since(t0))

	var wg sync.WaitGroup
	for threadID := 0; threadID < int(this.numShards); threadID++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			for i := 0; i < len(transitions); i++ {
				if transitions[i] != nil {
					if ids[i] == uint32(threadID) {
						this.sharded[threadID][keys[i]] = transitions[i].(UnivalueInterface).Value()
					}
				}
			}
		}(threadID)
	}
	wg.Wait()
}

func (this *DataStore) Print() {
	keys := []string{}
	for _, shard := range this.sharded {
		for k, v := range shard {
			if v != nil {
				keys = append(keys, k)
			}
		}
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return bytes.Compare([]byte(keys[i][:]), []byte(keys[j][:])) < 0
	})

	for _, key := range keys {
		fmt.Println("Key: ", key)
		this.sharded[key[0]][key].(TypeInterface).Print()
	}
}
