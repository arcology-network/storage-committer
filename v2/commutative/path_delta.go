package commutative

import (
	orderedset "github.com/arcology-network/common-lib/container/set"
)

type PathDelta struct {
	addDict *orderedset.OrderedSet
	delDict *orderedset.OrderedSet
}

func NewPathDelta(add []string, del []string) *PathDelta {
	return &PathDelta{
		orderedset.NewOrderedSet(add),
		orderedset.NewOrderedSet(del),
	}
}

func (this *PathDelta) Clone() *PathDelta {
	if this == nil {
		return this
	}
	return &PathDelta{
		this.addDict.Clone(),
		this.delDict.Clone(),
	}
}

func (this *PathDelta) Equal(other *PathDelta) bool {
	return this.addDict.Equal(other.addDict) &&
		this.delDict.Equal(other.delDict)
}

func (this *PathDelta) ProcessKey(subkey string, value interface{}, preexists bool) bool {
	return this.addKey(subkey, value, preexists) ||
		this.delKeys(subkey, value, preexists)
}

func (this *PathDelta) addKey(subkey string, value interface{}, preexists bool) bool {
	if this.delDict.Exists(subkey) { // Adding back a preexisting entry
		this.delDict.Delete(subkey) // Cancel out each other
		return true
	}

	if !preexists && value != nil {
		this.addDict.Insert(subkey) // won't be duplicate
		return true
	}
	return false
}

// Only the preexisting keys are in this buffer, or they will be cancel each other
func (this *PathDelta) delKeys(subkey string, value interface{}, preexists bool) bool {
	if value != nil {
		return false
	}

	if preexists { // Preexists and value == nil
		this.delDict.Insert(subkey)
		return true
	}
	return this.addDict.Delete(subkey) // Leave out the entry if it is in the added buffer
}

func (this *PathDelta) Added() interface{}   { return this.addDict.Keys() }
func (this *PathDelta) Removed() interface{} { return this.delDict.Keys() }
