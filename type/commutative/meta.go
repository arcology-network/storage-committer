package commutative

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	codec "github.com/HPISTechnologies/common-lib/codec"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/common"
)

type Meta struct {
	path      string // current path
	keys      []string
	added     []string // added keys
	removed   []string // removed keys
	finalized bool
}

func NewMeta(path string) (interface{}, error) {
	if !ccurlcommon.IsPath(path) {
		return nil, errors.New("Error: Wrong path format !")
	}

	if !ccurlcommon.CheckDepth(path) {
		return nil, errors.New("Error: Exceeded the maximum depth")
	}

	this := &Meta{
		path:      path,
		keys:      []string{},
		added:     []string{},
		removed:   []string{},
		finalized: false,
	}
	return this, nil
}

func (this *Meta) GobEncode() ([]byte, error) {

	return this.Encode(), nil
}
func (this *Meta) GobDecode(data []byte) error {
	meta := this.Decode(data).(*Meta)
	this.added = meta.added
	this.finalized = meta.finalized
	this.keys = meta.keys
	this.path = meta.path
	this.removed = meta.removed
	return nil
}
func (this *Meta) Deepcopy() interface{} {
	return &Meta{
		path:      this.path,
		keys:      this.keys,
		added:     this.added,
		removed:   this.removed,
		finalized: this.finalized,
	}
}

func (this *Meta) Value() interface{} {
	return this.keys
}

func (this *Meta) ToAccess() interface{} {
	return this
}

func (this *Meta) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	this.finalized = true
	this.Transitional(source.(*map[string]ccurlcommon.UnivalueInterface))
	if !this.Update(this.added, this.removed) {
		return this, 1, 0
	}

	this.added = []string{}
	this.removed = []string{}
	return this, 1, 1
}

func (this *Meta) Transitional(source interface{}) interface{} {
	this.added, this.removed = this.GetDifference(this.path, source.(*map[string]ccurlcommon.UnivalueInterface))
	return this
}

func (this *Meta) Set(tx uint32, path string, value interface{}, source interface{}) (uint32, uint32, error) {
	if value == nil {
		indexer := source.(ccurlcommon.DataSourceInterface)
		univalue := indexer.Read(tx, path)
		for _, subpath := range univalue.(*Meta).GetKeys() {
			indexer.Write(tx, path+subpath, nil) // Remove all the sub paths
		}
		return 0, 1, nil
	}
	return 0, 1, errors.New("Error: Path can only be created or deleted !")
}

func (this *Meta) Merge(tx uint32, other interface{}) {
	if other != nil && other.(ccurlcommon.TypeInterface).TypeID() == this.TypeID() {
		this.added = append(this.added, other.(*Meta).Added()...) // accumulate changes
		this.removed = append(this.removed, other.(*Meta).Removed()...)
	}
}

func (this *Meta) Composite() bool { return !this.finalized }
func (this *Meta) Finalize() {
	this.Update(this.added, this.removed)
	this.added = []string{}
	this.removed = []string{}
}

func (this *Meta) Path() string          { return this.path }
func (this *Meta) SetKeys(keys []string) { this.keys = keys }
func (this *Meta) GetKeys() []string     { return this.keys }
func (this *Meta) Added() []string       { return this.added }
func (this *Meta) Removed() []string     { return this.removed }
func (this *Meta) Updated() bool         { return len(this.added) > 0 || len(this.removed) > 0 }
func (this *Meta) TypeID() uint8         { return ccurlcommon.NoncommutativeMeta }

// Update subpaths
func (this *Meta) GetDifference(path string, cache *map[string]ccurlcommon.UnivalueInterface) ([]string, []string) {
	/* Always reprocess sub directories */
	addedKeys := make(map[string]bool)
	removedKeys := make(map[string]bool)

	for k, v := range *cache { // search for sub paths
		minLen := int(math.Min(float64(len(k)), float64(len(path))))
		if len(path) >= len(k) || k[:minLen] != path[:minLen] {
			continue
		}

		parts := strings.Split(k[len(path):], "/")
		if len(parts) > 1 && !(len(parts) == 2 && len(parts[1]) == 0) {
			continue
		}

		current := parts[0]
		if len(parts) > 1 {
			current += "/"
		}

		if v.GetWrites() == 0 { // ignore read only / create then delete
			continue
		}

		if !v.GetPreexist() && v.GetValue() != nil {
			addedKeys[current] = true
		}
		if v.GetValue() == nil {
			removedKeys[current] = true
		}
	}
	return ccurlcommon.GetMapKeys(&addedKeys), ccurlcommon.GetMapKeys(&removedKeys)
}

func (this *Meta) Update(addedKeys []string, removedKeys []string) bool {
	if len(addedKeys) == 0 && len(removedKeys) == 0 {
		return false
	}

	this.keys = ccurlcommon.RemoveFrom(ccurlcommon.Unique(append(this.keys, addedKeys...)), removedKeys)
	sort.SliceStable(this.keys, func(i, j int) bool {
		return strings.Compare(this.keys[i], this.keys[j]) < 0
	})

	return true
}

// Get all sub paths
func (this *Meta) GetSubpaths(tx uint32, path string, source interface{}) ([]string, error) {
	indexer := source.(ccurlcommon.DataSourceInterface)
	if !ccurlcommon.IsPath(path) {
		return []string{}, errors.New("Error: Invalid path format !")
	}

	subpaths := this.searcher(tx, path, indexer)
	sort.SliceStable(subpaths, func(i, j int) bool {
		return strings.Compare(subpaths[i], subpaths[j]) < 0
	})
	return subpaths, nil
}

func (this *Meta) searcher(tx uint32, path string, source interface{}) []string {
	indexer := source.(ccurlcommon.DataSourceInterface)
	subpaths := []string{}

	univalue := indexer.Read(tx, path)
	for _, current := range univalue.(*Meta).GetKeys() {
		subpaths = append(subpaths, path+current)
		if ccurlcommon.IsPath(path + current) {
			subpaths = append(subpaths, this.searcher(tx, path+current, indexer)...)
		}
	}
	return subpaths
}

func (this *Meta) Purge() {
	this.added = []string{}
	this.removed = []string{}
	this.finalized = false
}

func (this *Meta) Encode() []byte {
	byteset := [][]byte{
		codec.String(this.path).Encode(),
		codec.Strings(this.keys).Encode(),
		codec.Strings(this.added).Encode(),
		codec.Strings(this.removed).Encode(),
		codec.Bool(this.finalized).Encode(),
	}
	return codec.Byteset(byteset).Encode()
}

func (*Meta) Decode(bytes []byte) interface{} {
	fields := codec.Byteset{}.Decode(bytes)
	return &Meta{
		path:      codec.String("").Decode(fields[0]),
		keys:      codec.Strings([]string{}).Decode(fields[1]),
		added:     codec.Strings([]string{}).Decode(fields[2]),
		removed:   codec.Strings([]string{}).Decode(fields[3]),
		finalized: bool(codec.Bool(true).Decode(fields[4])),
	}
}

func (this *Meta) EncodeStripped() []byte {
	byteset := [][]byte{
		codec.String(this.path).Encode(),
		codec.Strings(this.keys).Encode(),
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Meta) DecodeStripped(bytes []byte) interface{} {
	fields := codec.Byteset{}.Decode(bytes)
	return &Meta{
		path:      codec.String("").Decode(fields[0]),
		keys:      codec.Strings([]string{}).Decode(fields[1]),
		added:     []string{},
		removed:   []string{},
		finalized: false,
	}
}

func (this *Meta) Print() {
	fmt.Println("Path: ", this.path)
	fmt.Println("Keys: ", this.keys)
	fmt.Println("Added: ", this.added)
	fmt.Println("Removed: ", this.removed)
	fmt.Println()
}
