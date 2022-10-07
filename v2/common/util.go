package common

import (
	"sort"
	"strings"
)

func IsPath(path string) bool {
	return !(len(path) == 0 || path[len(path)-1] != '/')
}

func CheckDepth(path string) bool {
	return strings.Count(path, "/") <= int(MaxDepth)
}

func GetParentPath(key string) string {
	if len(key) == 0 || key == Root {
		return key
	}
	path := key[:strings.LastIndex(key[:len(key)-1], "/")+1]
	return path
}

func RemoveFrom(from []string, toRemove []string) []string {
	removalLookup := make(map[string]bool, len(toRemove))
	for _, v := range toRemove {
		removalLookup[v] = true
	}

	target := make([]string, 0, len(from))
	for _, v := range from {
		if _, ok := removalLookup[v]; !ok {
			target = append(target, v)
		}
	}
	return target
}

func GetMapKeys(strMap *map[string]bool) []string {
	keys := []string{}
	for k := range *strMap { // search for sub paths
		keys = append(keys, k)
	}
	return keys
}

func ArrayToMap(keys []string) *map[string]bool {
	keyMap := make(map[string]bool)
	for _, k := range keys { // search for sub paths
		keyMap[k] = true
	}
	return &keyMap
}

func SortString(strings []string) []string {
	sort.SliceStable(strings, func(i, j int) bool {
		return strings[i] < strings[j]
	})
	return strings
}

func Deepcopy(keys []string) []string {
	cpy := make([]string, len(keys))
	copy(cpy, keys)
	return cpy
}

func Exclude(source []uint32, toRemove []uint32) []uint32 {
	txDict := make(map[uint32]bool)
	for _, tx := range source {
		txDict[tx] = true
	}

	for _, tx := range toRemove {
		delete(txDict, tx)
	}

	result := []uint32{}
	for tx := range txDict {
		result = append(result, tx)
	}
	return result
}

func RemoveNils(values *[]UnivalueInterface) {
	pos := int64(-1)
	for i := 0; i < len((*values)); i++ {
		if pos < 0 && (*values)[i] == nil {
			pos = int64(i)
			continue
		}

		if pos < 0 && (*values)[i] != nil {
			continue
		}

		if pos >= 0 && (*values)[i] == nil {
			(*values)[pos] = (*values)[i]
			continue
		}

		(*values)[pos] = (*values)[i]
		pos++
	}

	if pos >= 0 {
		(*values) = (*values)[:pos]
	}
}
