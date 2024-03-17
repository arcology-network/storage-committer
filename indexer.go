/*
*   Copyright (c) 2024 Arcology Network

*   This program is free software: you can redistribute it and/or modify
*   it under the terms of the GNU General Public License as published by
*   the Free Software Foundation, either version 3 of the License, or
*   (at your option) any later version.

*   This program is distributed in the hope that it will be useful,
*   but WITHOUT ANY WARRANTY; without even the implied warranty of
*   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
*   GNU General Public License for more details.

*   You should have received a copy of the GNU General Public License
*   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
package storagecommitter

import (
	"sync"

	"github.com/arcology-network/common-lib/exp/orderedmap"

	interfaces "github.com/arcology-network/storage-committer/interfaces"
)

type Indexer[K comparable, T, V any] struct {
	*orderedmap.OrderedMap[K, T, V]
	store    interfaces.Datastore
	lock     sync.Mutex
	ifAccept func(T) (K, bool) // If the transition is accepted and the key is returned. An index is only supposed to index the accepted transitions.
}

func NewIndexer[K comparable, T, V any](store interfaces.Datastore,
	ifAccept func(T) (K, bool),
	nilValue V,
	init func(K, T) V,
	setter func(K, T, *V)) *Indexer[K, T, V] {
	return &Indexer[K, T, V]{
		store:    store,
		ifAccept: ifAccept,

		OrderedMap: orderedmap.NewOrderedMap[K, T, V](
			nilValue,
			1024,
			init,
			setter,
		),
	}
}

// New creates a new StateCommitter instance.
func (this *Indexer[K, T, V]) Add(transitions []T) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, t := range transitions {
		if k, ok := this.ifAccept(t); ok {
			this.Set(k, t)
		}
	}
}

func (this *Indexer[K, T, V]) ParallelForeachDo(do func(k K, v *V)) {
	this.OrderedMap.ParallelForeachDo(do)
}
func (this *Indexer[K, T, V]) ForeachDo(do func(k K, v V)) { this.OrderedMap.ForeachDo(do) }
func (this *Indexer[K, T, V]) Clear()                      { this.OrderedMap.Clear() }
