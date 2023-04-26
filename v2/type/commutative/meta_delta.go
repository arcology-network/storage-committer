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

func (this *MetaDelta) Added() interface{}   { return this.addDict.Keys() }
func (this *MetaDelta) Removed() interface{} { return this.delDict.Keys() }
