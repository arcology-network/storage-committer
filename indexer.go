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
	"github.com/arcology-network/common-lib/exp/orderedmap"

	interfaces "github.com/arcology-network/storage-committer/interfaces"
)

type Indexer[K comparable, T any] struct {
	store    interfaces.Datastore
	index    *orderedmap.OrderedMap[K, T, []T]
	ifAccept func(T) (K, bool)
}

func NewIndexer[K comparable, T any](store interfaces.Datastore, ifAccept func(T) (K, bool)) *Indexer[K, T] {
	return &Indexer[K, T]{
		store:    store,
		ifAccept: ifAccept,

		index: orderedmap.NewOrderedMap[K, T, []T](
			nil,
			1024,
			func(k K, v T) []T {
				return []T{v}
			},
			func(k K, v T, seq *[]T) {
				*seq = append(*seq, v)
			}),
	}
}

// New creates a new StateCommitter instance.
func (this *Indexer[K, T]) Import(transitions []T) {
	for _, v := range transitions {
		if k, ok := this.ifAccept(v); ok {
			this.index.Set(k, v)
		}
	}
}
