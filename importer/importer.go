package importer

import (
	common "github.com/arcology-network/common-lib/common"
	ccmap "github.com/arcology-network/common-lib/exp/map"
	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/slice"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/interfaces"

	univalue "github.com/arcology-network/storage-committer/univalue"
	"github.com/cespare/xxhash/v2"
)

type Importer struct {
	numThreads int
	store      interfaces.Datastore
	platform   interfaces.Platform
	byTx       map[int]*[]*univalue.Univalue // Index by transaction id
	deltaDict  *ccmap.ConcurrentMap[string, *DeltaSequence]
	keyBuffer  []string      // Keys updated in the cycle
	valBuffer  []interface{} // Value updated in the cycle

	// seqPool *mempool.Mempool[*DeltaSequence] // Pool of sequences but very slow
}

func NewImporter(store interfaces.Datastore, platform interfaces.Platform) *Importer {
	var importer Importer
	importer.numThreads = 8
	importer.store = store

	importer.byTx = make(map[int]*[]*univalue.Univalue)
	// importer.byAcct = make(map[string]*[]*univalue.Univalue)

	importer.platform = platform
	importer.deltaDict = ccmap.NewConcurrentMap(8, func(v *DeltaSequence) bool { return v == nil }, func(k string) uint64 {
		return xxhash.Sum64([]byte(k))
		// return slice.Sum[uint8, uint8]([]byte(k))
	})

	return &importer
}

func (this *Importer) Init(store interfaces.Datastore) {
	this.store = store
	this.Clear()
}

func (this *Importer) SetStore(store interfaces.Datastore) { this.store = store }
func (this *Importer) Store() interfaces.Datastore         { return this.store }
func (this *Importer) ByPath() interface{}                 { return this.deltaDict }

func (this *Importer) IfExists(key string) bool {
	if _, ok := this.deltaDict.Get(key); !ok {
		return this.store.IfExists(key)
	}
	return false
}

// func (this *Importer) NewSequenceFromPool(k string, store interfaces.Datastore) *DeltaSequence {
// 	return this.seqPool.New().Init(k, store)
// }

// Remove entries that preexist but not available locally, it happens with a partial cache
func (this *Importer) RemoveRemoteEntries(txTrans *[]*univalue.Univalue, commitIfAbsent bool) {
	slice.RemoveIf(txTrans, func(_ int, univ *univalue.Univalue) bool {
		return univ.Preexist() && this.store.IfExists(*univ.GetPath()) && !commitIfAbsent
	})
}

func (this *Importer) Import(txTrans []*univalue.Univalue, args ...interface{}) []*DeltaSequence {
	commitIfAbsent := common.IfThenDo1st(len(args) > 0 && args[0] != nil, func() bool { return args[0].(bool) }, true) //Write if absent from local
	this.RemoveRemoteEntries(&txTrans, commitIfAbsent)

	// Create new sequences for the non-existing paths all at once.
	missingKeys := slice.ParallelTransform(txTrans, this.numThreads, func(i int, _ *univalue.Univalue) string {
		if _, ok := this.deltaDict.Get(*txTrans[i].GetPath()); !ok {
			return *txTrans[i].GetPath()
		}
		return ""
	})
	slice.Remove(&missingKeys, "") // Remove the keys that already exist.

	// Create the missing sequences as new transitions are being added.
	missingKeys = mapi.Keys(mapi.FromSlice(missingKeys, func(k string) bool { return true })) // Get the unique keys only by putting keys in a map

	newSeqs := slice.ParallelTransform(missingKeys, this.numThreads, func(i int, k string) *DeltaSequence {
		return NewDeltaSequence(k, this.store)
	})
	this.deltaDict.BatchSet(missingKeys, newSeqs)

	// Update the delta dictionary with the new sequences in parallel.
	this.deltaDict.ParallelFor(0, len(txTrans),
		func(i int) string { return *txTrans[i].GetPath() },
		func(i int, k string, seq *DeltaSequence, _ bool) (*DeltaSequence, bool) {
			return seq.Add(txTrans[i]), false
		})

	txIDs := slice.Transform(txTrans, func(_ int, v *univalue.Univalue) int { return int(v.GetTx()) })
	mapi.IfNotFoundDo(this.byTx, slice.UniqueInts(txIDs), func(k int) int { return k }, func(k int) *[]*univalue.Univalue {
		v := (make([]*univalue.Univalue, 0, 4)) // For unique ones only
		return &v
	})

	slice.ParallelForeach(txIDs, 4, func(i int, _ *int) {
		v := txTrans[i]
		tran := this.byTx[int(v.GetTx())]
		*tran = append(*tran, v)
	})

	return newSeqs
}

// Only keep transation within the whitelist
func (this *Importer) WhiteList(whitelist []uint32) []error {
	if whitelist == nil { // WhiteList all
		return []error{}
	}

	whitelisted := mapi.FromSlice(whitelist, func(_ uint32) bool { return true })

	for txid, vec := range this.byTx {
		if txid == stgcommcommon.SYSTEM {
			continue
		}

		if _, ok := whitelisted[uint32(txid)]; !ok {
			for _, v := range *vec {
				v.SetPath(nil) // Mark the transition status, so that it can be removed later.
			}
		}
	}
	return []error{}
}

func (this *Importer) SortDeltaSequences() {
	this.keyBuffer = this.deltaDict.Keys()

	slice.ParallelForeach(this.keyBuffer, this.numThreads, func(i int, _ *string) {
		deltaSeq, _ := this.deltaDict.Get(this.keyBuffer[i])
		deltaSeq.Sort() // Sort the transitions in the sequence
	})
}

// Merge and finalize state deltas
func (this *Importer) MergeStateDelta() {
	this.valBuffer = slice.Resize(&this.valBuffer, len(this.keyBuffer))

	slice.ParallelForeach(this.keyBuffer, this.numThreads, func(i int, _ *string) {
		deltaSeq, _ := this.deltaDict.Get(this.keyBuffer[i])
		this.valBuffer[i] = deltaSeq.Finalize()
	})

	slice.RemoveBothIf(&this.keyBuffer, &this.valBuffer, func(i int, _ string, v interface{}) bool {
		return this.valBuffer[i] == nil || this.valBuffer[i].(*univalue.Univalue) == nil
	})
}

func (this *Importer) KVs() ([]string, []interface{}) {
	slice.Remove(&this.keyBuffer, "")
	slice.RemoveIf(&this.valBuffer, func(_ int, v interface{}) bool { return v.(*univalue.Univalue) == nil })
	return this.keyBuffer, this.valBuffer
}

// Clear all
func (this *Importer) Clear() {
	for k, v := range this.byTx {
		*v = (*v)[:0]
		this.byTx[k] = v
	}

	clear(this.byTx)
	this.deltaDict.Clear()
	this.keyBuffer = this.keyBuffer[:0]
	this.valBuffer = this.valBuffer[:0]
	this.store.Clear()
}
