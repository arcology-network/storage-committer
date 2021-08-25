package common

import (
	"sort"
	"strings"
)

func IsPath(path string) bool {
	return !(len(path) == 0 || path[len(path)-1] != '/')
}

func CheckDepth(path string) bool {
	parts := strings.Split(path, "/")
	return len(parts) <= int(MaxDepth)
}

func GetLevel(path string) uint8 {
	return uint8(len(strings.Split(path, "/")))
}

func Unique(paths []string) []string {
	lookup := make(map[string]bool, len(paths))
	for _, v := range paths {
		lookup[v] = true
	}

	uniquePath := make([]string, 0, len(lookup))
	for v := range lookup {
		uniquePath = append(uniquePath, v)
	}

	sort.SliceStable(uniquePath, func(i, j int) bool {
		return uniquePath[i] < uniquePath[j]
	})
	return uniquePath
}

func IdxRange(start, end uint32) []uint32 {
	idxArr := make([]uint32, end-start)
	for i := start; i < end; i++ {
		idxArr[i-start] = i
	}
	return idxArr
}

func EqualLevel(lft, rgt string) bool {
	return len(strings.Split(lft, "/")) == len(strings.Split(rgt, "/"))
}

func GetParentPath(key string) string {
	if len(key) == 0 || key == Root {
		return key
	}

	parts := strings.Split(key, "/")
	if len(parts) == 1 {
		return key
	}

	if key[len(key)-1] == '/' { // a path
		return strings.Join(parts[:len(parts)-2], "/") + "/"
	}
	return strings.Join(parts[:len(parts)-1], "/") + "/"
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
