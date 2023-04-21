package ccdb

import (
	"fmt"
	"testing"
	"time"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	ccurltype "github.com/arcology-network/concurrenturl/v2/type"
	noncommutative "github.com/arcology-network/concurrenturl/v2/type/noncommutative"
	"github.com/arcology-network/concurrenturl/v2/univalue"
)

func TestCachePolicyLowScore(t *testing.T) {
	/*
		keys := make([]string, 5)
		nVals := make([]interface{}, len(keys))
		for i := 0; i < len(keys); i++ {
			keys[i] = fmt.Sprint(i)
			nVals[i] = univalue.NewUnivalue( uint32(i), keys[i], 1, 1, 2, noncommutative.NewInt64(int64(i)))
		}

		for i := 2; i < len(keys); i++ {
			keys[i] = fmt.Sprint(i)
			nVals[i] = univalue.NewUnivalue( uint32(i), keys[i], 2, 2, noncommutative.NewInt64(int64(i)))
		}

		persistentDB := cachedstorage.NewMemDB()
		cachePolicy := cachedstorage.NewCachePolicy(1, 1.0) // 1 byte only
		dataStore := cachedstorage.NewDataStore(nil, cachePolicy, persistentDB, ccurltype.ToBytes, ccurltype.FromBytes)

		// -------------------- First insertion --------------------
		dataStore.BatchInject(keys, nVals)
		//dataStore.CachePolicy().AddToBuffer(keys, nVals)
		dataStore.CachePolicy().Refresh(dataStore.LocalCache())
		dataStore.CachePolicy().PrintScores()

		entiresFreed, memFreed := dataStore.CachePolicy().Refresh(dataStore.LocalCache())
		if entiresFreed != 5 || dataStore.CachePolicy().Size() != 0 || dataStore.LocalCache().Size() != 0 {
			t.Error("dataStore.cachePolicy should be empty: ", entiresFreed, " dataStore.cachePolicy.scoreboard:", memFreed, " bytes")
		}

		// --------------Second data batch with higher scores---------------------
		for i := 0; i < len(keys); i++ {
			keys[i] = fmt.Sprint(i + 10)
			nVals[i] = univalue.NewUnivalue( uint32(i), keys[i], 129, 129, noncommutative.NewInt64(int64(i+10)))
		}

		// cachePolicy = cachedstorage.NewCachePolicy(1, 1.0) // 1 byte only
		// dataStore = cachedstorage.NewDataStore(nil, cachePolicy)

		// --------------Second insertion ---------------------
		dataStore.BatchInject(keys, nVals)
		//dataStore.CachePolicy().AddToBuffer(keys, nVals)
		dataStore.CachePolicy().Refresh(dataStore.LocalCache())

		entiresFreed, _ = dataStore.CachePolicy().Refresh(dataStore.LocalCache())
		if entiresFreed == 5 || dataStore.CachePolicy().Size() == 0 || dataStore.LocalCache().Size() == 0 {
			dataStore.CachePolicy().PrintScores()
			t.Error("Error: There should be 0 ", dataStore.CachePolicy().Size(), dataStore.LocalCache().Size())
		}
	*/
}

func TestCachePolicyAndPersistentDB(t *testing.T) {
	/*
		keys := make([]string, 10)
		nVals := make([]interface{}, len(keys))
		for i := 0; i < len(keys); i++ {
			keys[i] = fmt.Sprint(i)
			nVals[i] = univalue.NewUnivalue( uint32(i), keys[i], 1, 1, 2, noncommutative.NewInt64(int64(i)))
		}

		persistentDB := cachedstorage.NewMemDB()
		cachePolicy := cachedstorage.NewCachePolicy(1, 0.8)
		dataStore := cachedstorage.NewDataStore(nil, cachePolicy, persistentDB, ccurltype.ToBytes, ccurltype.FromBytes)

		// First insertion
		dataStore.BatchInject(keys, nVals)
		//dataStore.CachePolicy().AddToBuffer(keys, nVals)
		dataStore.CachePolicy().Refresh(dataStore.LocalCache())
		dataStore.CachePolicy().PrintScores()

		entiresFreed, memFreed := dataStore.CachePolicy().Refresh(dataStore.LocalCache())
		if entiresFreed != 10 || memFreed == 0 || dataStore.CachePolicy().Size() != 0 || dataStore.LocalCache().Size() != 0 {
			t.Error("dataStore.cachePolicy should be empty: ", entiresFreed, " dataStore.cachePolicy.scoreboard:", memFreed, " bytes")
		}

		//cachePolicy.AdjustThreshold(2000, 0.8)
		dataStore.BatchInject(keys, nVals)

		//dataStore.CachePolicy().AddToBuffer(keys, nVals)
		dataStore.CachePolicy().Refresh(dataStore.LocalCache())
		dataStore.CachePolicy().PrintScores()

		entiresFreed, memFreed = dataStore.CachePolicy().Refresh(dataStore.LocalCache())
		if entiresFreed == 0 || memFreed == 0 {
			fmt.Println("Error: Freed Entries: ", entiresFreed, " size: ", memFreed, " bytes")
			dataStore.CachePolicy().PrintScores()
		}
	*/
}

func TestCacheWithPersistentStorage(t *testing.T) {
	keys := make([]string, 5)
	encoded := make([][]byte, len(keys))
	nVals := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		keys[i] = fmt.Sprint(i)
		nVals[i] = univalue.NewUnivalue(uint32(i), keys[i], 1, 1, 2, noncommutative.NewInt64(int64(i)))
		encoded[i] = ccurltype.ToBytes(nVals[i].(ccurlcommon.UnivalueInterface).Value())
	}
	persistentDB := cachedstorage.NewMemDB()
	persistentDB.BatchSet(keys, encoded)
	cachePolicy := cachedstorage.NewCachePolicy(1, 0.8)
	dataStore := cachedstorage.NewDataStore(nil, cachePolicy, persistentDB, ccurltype.ToBytes, ccurltype.FromBytes)

	// First insertion
	dataStore.BatchInject(keys, nVals)
	//dataStore.CachePolicy().AddToBuffer(keys, nVals)
	dataStore.CachePolicy().Refresh(dataStore.LocalCache())
	dataStore.CachePolicy().PrintScores()

	dataStore.CachePolicy().Refresh(dataStore.LocalCache())
	dataStore.CachePolicy().PrintScores()

	v := dataStore.BatchRetrive(keys)
	for i := 0; i < len(v); i++ {
		if v[i] == nil {
			t.Error("Error: Should have retrived 5 nonnil entries !")
		}
	}

	if dataStore.CachePolicy().Size() != 0 {
		t.Error("Error: Should be empty !")
	}
}

func BenchmarkCachePolicy(b *testing.B) {
	keys := make([]string, 1000000)
	nVals := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		keys[i] = fmt.Sprint(i)
		nVals[i] = univalue.NewUnivalue(uint32(i), keys[i], 1, 1, 2, noncommutative.NewInt64(int64(i)))
	}

	cachePolicy := cachedstorage.NewCachePolicy(10, 0.8)
	dataStore := cachedstorage.NewDataStore(nil, cachePolicy)
	dataStore.BatchInject(keys, nVals)

	t0 := time.Now()
	//dataStore.CachePolicy().AddToBuffer(keys, nVals)
	dataStore.CachePolicy().Refresh(dataStore.LocalCache())
	fmt.Println("CachePolicy Refresh:", time.Since(t0))

	t0 = time.Now()
	dataStore.CachePolicy().Refresh(dataStore.LocalCache())
	fmt.Println("CachePolicy FreeMemory:", time.Since(t0))
}
