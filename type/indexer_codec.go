package urltype

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	ccurlcommon "github.com/arcology/concurrenturl/common"
)

/* Export access records */
func (this *Indexer) ToArray() []ccurlcommon.UnivalueInterface {
	buffer := make([]ccurlcommon.UnivalueInterface, 0, len(this.buffer))
	for _, v := range this.buffer {
		buffer = append(buffer, v)
	}

	sort.SliceStable(buffer, func(i, j int) bool {
		return strings.Compare(buffer[i].GetPath(), buffer[j].GetPath()) < 0
	})
	return buffer
}

func (this *Indexer) Equal(other *Indexer) bool {
	cache0 := this.ToArray()
	cache1 := other.ToArray()
	cacheFlag := reflect.DeepEqual(cache0, cache1)
	return cacheFlag
}

func (this *Indexer) Print() {
	for i, elem := range this.ToArray() {
		fmt.Println("Level : ", i)
		elem.Print()
	}
}
