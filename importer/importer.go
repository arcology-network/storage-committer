package importer

import (
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/array"
	ccmap "github.com/arcology-network/common-lib/exp/map"
	mapi "github.com/arcology-network/common-lib/exp/map"
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

	acctBuffer     []string
	acctTranBuffer []*[]*univalue.Univalue

	// seqPool *mempool.Mempool[*DeltaSequence] // Pool of sequences but very slow
}

func NewImporter(store interfaces.Datastore, platform interfaces.Platform) *Importer {
	var importer Importer
	importer.numThreads = 8
	importer.store = store

	importer.byTx = make(map[int]*[]*univalue.Univalue)
	// importer.byAcct = make(map[string]*[]*univalue.Univalue)

	importer.acctBuffer = []string{}
	importer.acctTranBuffer = []*[]*univalue.Univalue{}

	importer.platform = platform
	importer.deltaDict = ccmap.NewConcurrentMap(8, func(v *DeltaSequence) bool { return v == nil }, func(k string) uint64 {
		return xxhash.Sum64([]byte(k))
		// return array.Sum[uint8, uint8]([]byte(k))
	})

	// importer.seqPool = mempool.NewMempool[*DeltaSequence](4096, 64, func() *DeltaSequence {
	// 	return NewDeltaSequence("", nil) // Init an empty sequence.
	// }, func(_ *DeltaSequence) {})
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

func (this *Importer) Import(txTrans []*univalue.Univalue, args ...interface{}) []*DeltaSequence {
	commitIfAbsent := common.IfThenDo1st(len(args) > 0 && args[0] != nil, func() bool { return args[0].(bool) }, true) //Write if absent from local

	//Remove entries that preexist but not available locally, it happens with a partial cache
	array.RemoveIf(&txTrans, func(_ int, univ *univalue.Univalue) bool {
		return univ.Preexist() && this.store.IfExists(*univ.GetPath()) && !commitIfAbsent
	})

	// Create new sequences for the non-existing paths all at once.
	missingKeys := array.ParallelAppend(txTrans, this.numThreads, func(i int, _ *univalue.Univalue) string {
		if _, ok := this.deltaDict.Get(*txTrans[i].GetPath()); !ok {
			return *txTrans[i].GetPath()
		}
		return ""
	})
	array.Remove(&missingKeys, "") // Remove the keys that already exist.

	// Create the missing sequences as new transitions are being added.
	missingKeys = mapi.Keys(mapi.FromSlice(missingKeys, func(k string) bool { return true })) // Get the unique keys only

	newSeqs := array.ParallelAppend(missingKeys, this.numThreads, func(i int, k string) *DeltaSequence {
		return NewDeltaSequence(k, this.store)
	})
	this.deltaDict.BatchSet(missingKeys, newSeqs)

	// Update the delta dictionary with the new sequences in parallel.
	this.deltaDict.ParallelFor(0, len(txTrans),
		func(i int) string { return *txTrans[i].GetPath() },
		func(i int, k string, seq *DeltaSequence, _ bool) (*DeltaSequence, bool) {
			return seq.Add(txTrans[i]), false
		})

	txIDs := array.Append(txTrans, func(_ int, v *univalue.Univalue) int { return int(v.GetTx()) })
	mapi.IfNotFoundDo(this.byTx, array.UniqueInts(txIDs), func(k int) int { return k }, func(k int) *[]*univalue.Univalue {
		v := (make([]*univalue.Univalue, 0, 16)) // For unique ones only
		return &v
	})

	array.ParallelForeach(txIDs, 4, func(i int, _ *int) {
		v := txTrans[i]
		tran := this.byTx[int(v.GetTx())]
		*tran = append(*tran, v)
	})

	return newSeqs

	// Create an array of transactions for each account.
	// mapi.IfNotFoundDo(this.byAcct, txTrans,
	// 	func(univ *univalue.Univalue) string { return platform.GetAccountAddr(*univ.GetPath()) },
	// 	func(_ string) *[]*univalue.Univalue {
	// 		return common.New(make([]*univalue.Univalue, 0, 8))
	// 	})

	// // Add the transaction to the account's array.
	// array.Foreach(txTrans, func(i int, v **univalue.Univalue) {
	// 	tran := this.byAcct[platform.GetAccountAddr(*(*v).GetPath())]
	// 	*tran = append(*tran, *v)
	// })
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

	array.ParallelForeach(this.keyBuffer, this.numThreads, func(i int, _ *string) {
		deltaSeq, _ := this.deltaDict.Get(this.keyBuffer[i])
		deltaSeq.Sort() // Sort the transitions in the sequence
	})
}

// Merge and finalize state deltas
func (this *Importer) MergeStateDelta() {
	this.valBuffer = array.Resize(&this.valBuffer, len(this.keyBuffer))

	array.ParallelForeach(this.keyBuffer, this.numThreads, func(i int, _ *string) {
		deltaSeq, _ := this.deltaDict.Get(this.keyBuffer[i])
		this.valBuffer[i] = deltaSeq.Finalize()
	})

	array.RemoveBothIf(&this.keyBuffer, &this.valBuffer, func(i int, _ string, v interface{}) bool {
		return this.valBuffer[i] == nil || this.valBuffer[i].(*univalue.Univalue) == nil
	})
}

func (this *Importer) KVs() ([]string, []interface{}) {
	array.Remove(&this.keyBuffer, "")
	array.RemoveIf(&this.valBuffer, func(_ int, v interface{}) bool { return v.(*univalue.Univalue) == nil })
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

	this.acctBuffer = this.acctBuffer[:0]
	this.acctTranBuffer = this.acctTranBuffer[:0]
	// t0 := time.Now()
	// this.seqPool.Reset() // Very slow
	// this.seqPool.ReclaimRecursive()
	// fmt.Println("Reset: ", time.Since(t0))
	this.store.Clear()
}
