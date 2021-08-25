package commutative

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	codec "github.com/HPISTechnologies/common-lib/codec"
	ccurlcommon "github.com/HPISTechnologies/concurrenturl/v2/common"
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

func (this *Meta) Deepcopy() interface{} {
	return &Meta{
		path:      this.path,
		keys:      ccurlcommon.Deepcopy(this.keys),
		added:     ccurlcommon.Deepcopy(this.added),
		removed:   ccurlcommon.Deepcopy(this.removed),
		finalized: this.finalized,
	}
}

func (this *Meta) Value() interface{} {
	return this.keys
}

func (this *Meta) ToAccess() interface{} {
	return nil
}

func (this *Meta) Get(tx uint32, path string, source interface{}) (interface{}, uint32, uint32) {
	this.finalized = true
	temp := &Meta{ // make a temporary copy
		path:      this.path,
		keys:      ccurlcommon.Deepcopy(this.keys),
		added:     []string{},
		removed:   []string{},
		finalized: this.finalized,
	}

	delta := this.Delta(source.(*map[string]ccurlcommon.UnivalueInterface)).(*Meta)
	if !temp.Update(delta.Added(), delta.Removed()) {
		return this, 1, 0
	}

	return temp, 1, 1
}

func (this *Meta) Peek(source interface{}) interface{} {
	tempCopy := this.Deepcopy().(*Meta)
	tempCopy.Get(ccurlcommon.SYSTEM, this.path, source)
	return tempCopy
}

func (this *Meta) Delta(source interface{}) interface{} {
	path := this.path
	cache := source.(*map[string]ccurlcommon.UnivalueInterface)
	added := make([]string, 0, len(*cache))
	removed := make([]string, 0, len(*cache))

	for k, v := range *cache { // search for sub paths
		if v.GetWrites() == 0 { // ignore read only / create then delete
			continue
		}

		minLen := int(math.Min(float64(len(k)), float64(len(path))))
		if len(path) >= len(k) || k[:minLen] != path[:minLen] {
			continue
		}

		pos := ccurlcommon.Find(k, len(path), len(k), '/')
		if pos >= 0 && len(k) != pos+1 {
			continue
		}

		current := k[len(path):]
		if !v.GetPreexist() && v.GetValue() != nil {
			added = append(added, current)
		}

		if v.GetValue() == nil {
			removed = append(removed, current)
		}
	}

	delta := &Meta{ // make a temporary copy
		path:      this.path,
		keys:      []string{},
		added:     added,
		removed:   removed,
		finalized: this.finalized,
	}
	return delta
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

func (this *Meta) ApplyDelta(tx uint32, other interface{}) {
	if other != nil && other.(ccurlcommon.TypeInterface).TypeID() == this.TypeID() {
		this.added = append(this.added, other.(*Meta).Added()...) // accumulate changes
		this.removed = append(this.removed, other.(*Meta).Removed()...)
	}
}

func (this *Meta) Finalize() {
	this.Update(this.added, this.removed)
	this.added = []string{}
	this.removed = []string{}
}

func (this *Meta) Composite() bool       { return !this.finalized }
func (this *Meta) Path() string          { return this.path }
func (this *Meta) SetKeys(keys []string) { this.keys = keys }
func (this *Meta) GetKeys() []string     { return this.keys }
func (this *Meta) Added() []string       { return this.added }
func (this *Meta) Removed() []string     { return this.removed }
func (this *Meta) Updated() bool         { return len(this.added) > 0 || len(this.removed) > 0 }
func (this *Meta) TypeID() uint8         { return ccurlcommon.CommutativeMeta }

func (this *Meta) Update(addedKeys []string, removedKeys []string) bool {
	if len(addedKeys) == 0 && len(removedKeys) == 0 {
		return false
	}

	this.keys = ccurlcommon.RemoveFrom(ccurlcommon.Unique(append(this.keys, addedKeys...)), removedKeys) // 80
	sort.SliceStable(this.keys, func(i, j int) bool {
		return strings.Compare(this.keys[i], this.keys[j]) < 0
	})
	return true
}

func (this *Meta) Purge() {
	this.added = []string{}
	this.removed = []string{}
	this.finalized = false
}

func (this *Meta) Hash(hasher func([]byte) []byte) []byte {
	return hasher(this.EncodeCompact())
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

func (this *Meta) EncodeCompact() []byte {
	byteset := [][]byte{
		codec.String(this.path).Encode(),
		codec.Strings(this.keys).Encode(),
	}
	return codec.Byteset(byteset).Encode()
}

func (this *Meta) DecodeCompact(bytes []byte) interface{} {
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
