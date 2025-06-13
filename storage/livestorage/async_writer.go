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

package ccstorage

// LiveStorageWriter is a struct that contains data strucuture and methods for writing data to concurrent storage.
type LiveStorageWriter struct {
	*LiveStgIndexer
	buffer  []*LiveStgIndexer
	store   *LiveStorage
	version int64
}

func NewLiveStorageWriter(store *LiveStorage, version int64) *LiveStorageWriter {
	return &LiveStorageWriter{
		LiveStgIndexer: NewLiveStgIndexer(store, 0),
		buffer:         []*LiveStgIndexer{},
		store:          store,
		version:        version,
	}
}

// Send the data to the downstream processor. This can be called multiple times
// before calling Await to commit the data to the state db.
func (this *LiveStorageWriter) Precommit(isSync bool) {
	if isSync {
		this.LiveStgIndexer.PreCommit()
	} else {
		this.LiveStgIndexer.Finalize() // Remove the nil transitions
		this.buffer = append(this.buffer, this.LiveStgIndexer)
		this.LiveStgIndexer = NewLiveStgIndexer(this.store, -1)
	}
}

// Await commits the data to the state db.
func (this *LiveStorageWriter) Commit(_ uint64) {
	mergedIdxer := new(LiveStgIndexer).Merge(this.buffer)
	var err error
	if this.store.db != nil {
		if err = this.store.db.BatchSet(mergedIdxer.keyBuffer, mergedIdxer.encodedBuffer); err != nil {
			panic(err)
		}
	}
	this.store.cache.BatchSet(mergedIdxer.keyBuffer, mergedIdxer.valueBuffer) // update the local cache
	this.buffer = this.buffer[:0]
}

func (this *LiveStorageWriter) Name() string { return "Live Storage Writer" }
