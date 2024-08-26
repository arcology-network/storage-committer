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

package cache

// ExecutionCacheWriter is a struct that contains data strucuture and methods for writing data to cache.
// The indexer is used to index the input transitions as they are received, in a way that they can be committed efficiently later.
type ExecutionCacheWriter struct {
	*WriteCacheIndexer
	*WriteCache
}

func NewExecutionCacheWriter(writeCache *WriteCache, version int64) *ExecutionCacheWriter {
	return &ExecutionCacheWriter{
		WriteCacheIndexer: NewWriteCacheIndexer(nil, int64(version)),
		WriteCache:        writeCache,
	}
}

// write cache updates itself every generation. It doesn't need to write to the database.
func (this *ExecutionCacheWriter) Precommit() {
	this.WriteCacheIndexer.Finalize() // Remove the nil transitions
	for i := range this.WriteCacheIndexer.buffer {
		this.WriteCache.kvDict[*this.WriteCacheIndexer.buffer[i].GetPath()] = this.WriteCacheIndexer.buffer[i]
	}
	this.WriteCacheIndexer = NewWriteCacheIndexer(nil, -1)

}

// The generation cache is transient and will clear itself when all the transitions are committed to
// the database.
func (this *ExecutionCacheWriter) Commit(_ uint64) {
	this.WriteCache.Clear()
	this.WriteCacheIndexer.buffer = this.WriteCacheIndexer.buffer[:0]
}
