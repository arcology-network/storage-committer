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
package livecache

import (
	"math"
	"sort"
	"sync/atomic"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/associative"
	"github.com/arcology-network/common-lib/exp/slice"
	stgtype "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/type/univalue"
)

type Usage struct {
	// key       *string
	sizeInMem uint64
	// lastLoaded  uint32
	// firstLoaded uint32
	visits uint64
}

type CacheUsage struct {
	liveCache *LiveCache
	minScore  float64
	maxScore  float64
	dist      [65536]uint64 // size distribution
	occupied  uint64        // The total memory used by the cache.
	maxSize   uint64
}

func NewCacheUsage(maxSize uint64, liveCache *LiveCache) *CacheUsage {
	usage := &CacheUsage{
		// lookup:   make(map[string]*Usage),
		liveCache: liveCache,
		minScore:  0,
		maxScore:  0,
		// keys:      paged.NewPagedSlice[*Usage](1024, 100, 0),
		dist:     [65536]uint64{},
		occupied: 0,
		maxSize:  uint64(24 * 1024 * 1024 * 1024), // 0.8 of the minimum memory required.
	}

	if v, err := common.GetAvailableMemory(); err == nil {
		usage.maxSize = common.Min(maxSize, uint64(float64(v)*0.8))
	}
	return usage
}

func (this *CacheUsage) MaxScore() float64 { return this.maxScore }
func (this *CacheUsage) MinScore() uint64  { return this.occupied }

// Update updates the cache usage statistics.
func (this *CacheUsage) UpdateStats(univals []*univalue.Univalue) {
	for _, univ := range univals {
		if pair, _ := this.liveCache.GetRaw(*univ.GetPath()); pair != nil { // Exists in the cache.
			oldBin := math.Round(float64(pair.Second.visits) / float64(math.MaxUint32/uint32(len(this.dist))))
			this.dist[uint32(oldBin)] -= pair.Second.sizeInMem
		}

		if univ.Value() == nil {
			continue
		}

		visits := uint64(univ.Reads() + univ.Writes() + univ.DeltaWrites())
		newBin := math.Round(float64(visits) / float64(math.MaxUint32/uint32(len(this.dist))))
		this.dist[uint32(newBin)] += univ.Value().(stgtype.Type).MemSize()
	}
}

// Check if the cache has enough space to store the new values.
// If not, the cache will be cleared. If still not enough space,
// some new values won't be stored.
func (this *CacheUsage) PrepareSpace(univals *[]*univalue.Univalue, liveCache *LiveCache) (uint64, error) {
	// The total memory required to store the new values.
	totalRequired := slice.Accumulate[*univalue.Univalue, uint64](*univals, 0, func(_ int, v *univalue.Univalue) uint64 {
		if v.Value() == nil {
			return 0
		}
		return v.Value().(stgtype.Type).MemSize()
	})

	// The available memory to store the new values.
	availableMemory, err := common.GetAvailableMemory()
	if err != nil {
		return 0, err
	}
	actualCap := common.Min(this.maxSize, availableMemory) // The actual cap of the cache.

	// The memory that needs to be freed to store the new values.
	toFree := int(totalRequired) - (int(actualCap) - int(this.occupied))
	if toFree <= 0 {
		this.occupied = totalRequired
		return 0, nil // Enough space, no need to free memory.
	}

	// Check if the cache has enough space to store the new values.
	// If not, remove the values.
	freedMemory := this.freeCache(uint64(toFree))

	// Not enough memory for all even after freeing some memory. Some new values won't be stored.
	// Sort the univalues by size in memory, so the smallest values will still have a chance to be stored.
	if totalRequired > availableMemory+freedMemory {
		sort.Slice(*univals, func(i, j int) bool {
			return (*univals)[i].Value().(stgtype.Type).MemSize() < (*univals)[j].Value().(stgtype.Type).MemSize()
		})

		// Find the index of the last value that can be stored in the cache.
		idx := 0
		accumSize := uint64(0)
		for i, v := range *univals {
			accumSize += v.Value().(stgtype.Type).MemSize()
			if accumSize > availableMemory+freedMemory {
				idx = i
				break
			}
		}

		// Accumulated size minus the last value's size which it isn't stored in the cache.
		// Because it made the accumulated size exceed the available memory.
		this.occupied = accumSize - (*univals)[idx].Value().(stgtype.Type).MemSize()
		*univals = (*univals)[:idx] // Some new values won't be stored in cache.
		return freedMemory, nil
	}

	this.occupied += totalRequired - freedMemory
	return freedMemory, nil
}

// freeCache frees the required memory.
func (this *CacheUsage) freeCache(sizeToFree uint64) uint64 {
	bin := 0
	for i := 0; i < len(this.dist); i++ {
		if sizeToFree -= this.dist[i]; sizeToFree <= 0 {
			bin = i
		}
	}
	threshold := uint64(bin+1) * uint64(math.MaxUint32/uint32(len(this.dist)))

	// delete the values from the cache and accumulate the freed memory.
	var totalFreed atomic.Uint64
	this.liveCache.ConcurrentMap.ParallelDelete(func(_ string, v *associative.Pair[stgtype.Type, *Usage]) bool {
		if uint64(math.Pow(float64(v.Second.visits), math.E)) <= threshold {
			totalFreed.Add(v.Second.sizeInMem)
			return true
		}
		return false
	})
	return totalFreed.Load()
}
