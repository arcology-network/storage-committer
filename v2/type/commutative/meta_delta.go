package commutative

import (
	orderedset "github.com/arcology-network/common-lib/container/set"
)

type MetaDelta struct {
	addDict *orderedset.OrderedSet
	delDict *orderedset.OrderedSet
}

func NewMetaDelta(add []string, del []string) *MetaDelta {
	return &MetaDelta{
		orderedset.NewOrderedSet(add),
		orderedset.NewOrderedSet(del),
	}
}

func (this *MetaDelta) Clone() *MetaDelta {
	if this == nil {
		return this
	}
	return &MetaDelta{
		this.addDict.Clone(),
		this.delDict.Clone(),
	}
}

func (this *MetaDelta) ProcessKey(subkey string, value interface{}, preexists bool) bool {
	return this.addKey(subkey, value, preexists) ||
		this.delKeys(subkey, value, preexists)
}

func (this *MetaDelta) addKey(subkey string, value interface{}, preexists bool) bool {
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
func (this *MetaDelta) delKeys(subkey string, value interface{}, preexists bool) bool {
	if value != nil {
		return false
	}

	if preexists { // Preexists and value == nil
		this.delDict.Insert(subkey)
		return true
	}
	return this.addDict.Delete(subkey) // Leave out the entry if it is in the added buffer
}

func (this *MetaDelta) Added() interface{}   { return this.addDict.Keys() }
func (this *MetaDelta) Removed() interface{} { return this.delDict.Keys() }
