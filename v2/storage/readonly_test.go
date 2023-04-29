package ccdb

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	noncommutative "github.com/arcology-network/concurrenturl/v2/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/v2/univalue"
)

func TestReadonlyStorageLocal(t *testing.T) {
	// Server end
	persistentDB := cachedstorage.NewMemDB()
	serverCachePolicy := cachedstorage.NewCachePolicy(1, 0.8)
	serverDataStore := cachedstorage.NewDataStore(nil, serverCachePolicy, persistentDB, univalue.ToBytes, univalue.FromBytes)

	keys := []string{}
	values := []interface{}{}
	for i := 0; i < 8; i++ { // 8 in the server db
		keys = append(keys, fmt.Sprint(i))
		v := univalue.NewUnivalue(uint32(i), fmt.Sprint(i), 1, 1, 2, noncommutative.NewInt64(int64(i)))
		values = append(values, v)

		persistentDB.Set(fmt.Sprint(i), univalue.ToBytes(noncommutative.NewInt64(int64(i)))) // save to the DB directly
	}
	serverDataStore.Precommit(keys[:4], values[:4]) // 4 in the server side cache
	serverDataStore.Commit()

	// Simulated Client
	keys1 := []string{}
	values1 := []interface{}{}
	for i := 0; i < 8; i++ { // 8 in the server db
		keys1 = append(keys1, fmt.Sprint(i))
		values1 = append(values1, univalue.NewUnivalue(uint32(i), fmt.Sprint(i), 1, 1, 2, noncommutative.NewInt64(int64(i))))
	}

	placeholderEncoder := func(v interface{}) []byte { return univalue.ToBytes(v) }
	placeholderDecoder := func(bytes []byte) interface{} { return univalue.FromBytes(bytes) }

	readonlyClientProxy := NewReadonlyClient("", "", nil, serverDataStore)
	clientCachePolicy := cachedstorage.NewCachePolicy(1, 0.8)
	clientDataStore := cachedstorage.NewDataStore(nil, clientCachePolicy, readonlyClientProxy, placeholderEncoder, placeholderDecoder)
	clientDataStore.Precommit(keys1[:2], values1[:2]) // 2 in the client side cache
	clientDataStore.Commit()

	// Retrive two entries from the client cache
	v0, err := clientDataStore.Retrive(keys1[0])
	if err != nil {
		t.Errorf("Retrive Error: %v path=%v !", err, keys1[0])
	}
	v1, err := clientDataStore.Retrive(keys1[1])
	if err != nil {
		t.Errorf("Retrive Error: %v path=%v !", err, keys1[1])
	}
	if v0 != values1[0] || v1 != values1[1] {
		t.Errorf("Error: Failed to retrive entries from client cache !")
	}
	v2, err := clientDataStore.Retrive(keys1[2])
	if err != nil {
		t.Errorf("Retrive Error: %v path=%v !", err, keys1[2])
	}
	v3, err := clientDataStore.Retrive(keys1[3])
	if err != nil {
		t.Errorf("Retrive Error: %v path=%v !", err, keys1[3])
	}
	if v2 == nil || v3 == nil {
		t.Error("Error: Failed to retrive entries from client cache !")
	}
	//readonlyClientProxy
}

func TestReadonlyStorageRemote(t *testing.T) {
	// Server end
	persistentDB := cachedstorage.NewMemDB()
	serverCachePolicy := cachedstorage.NewCachePolicy(1, 0.8)
	serverDataStore := cachedstorage.NewDataStore(nil, serverCachePolicy, persistentDB, univalue.ToBytes, univalue.FromBytes)

	keys := []string{}
	values := []interface{}{}
	for i := 0; i < 8; i++ { // 8 in the server db
		keys = append(keys, fmt.Sprint(i))
		v := univalue.NewUnivalue(uint32(i), fmt.Sprint(i), 1, 1, 2, noncommutative.NewInt64(int64(i)))
		values = append(values, v)
		persistentDB.Set(fmt.Sprint(i), univalue.ToBytes(noncommutative.NewInt64(int64(i)))) // save to the DB directly
	}
	serverDataStore.Precommit(keys[:4], values[:4]) // 4 in the server side cache
	serverDataStore.Commit()

	server := NewReadonlyServer("", univalue.ToBytes, univalue.FromBytes, serverDataStore)
	go func() {
		http.HandleFunc("/store", server.Receive)
		http.ListenAndServe(":8090", nil)
	}()
	time.Sleep(5 * time.Second)

	keys1 := []string{}
	values1 := []interface{}{}
	for i := 0; i < 8; i++ { // 8 in the server db
		keys1 = append(keys1, fmt.Sprint(i))
		values1 = append(values1, univalue.NewUnivalue(uint32(i), fmt.Sprint(i), 1, 1, 2, noncommutative.NewInt64(int64(i))))
	}

	proxyEncoder := func(v interface{}) []byte { return univalue.ToBytes(v) }
	proxyDecoder := func(bytes []byte) interface{} { return univalue.FromBytes(bytes) }

	readonlyClientProxy := NewReadonlyClient("http://localhost:8090", "store", nil)
	clientCachePolicy := cachedstorage.NewCachePolicy(1, 0.8)
	clientDataStore := cachedstorage.NewDataStore(nil, clientCachePolicy, readonlyClientProxy, proxyEncoder, proxyDecoder)
	clientDataStore.Precommit(keys1[:2], values1[:2]) // 2 in the client side cache
	clientDataStore.Commit()

	// Retrive two entries from the client cache
	v0, err := clientDataStore.Retrive(keys1[0])
	if err != nil {
		t.Errorf("Retrive Error: %v path=%v !", err, keys1[0])
	}
	v1, err := clientDataStore.Retrive(keys1[1])
	if err != nil {
		t.Errorf("Retrive Error: %v path=%v !", err, keys1[1])
	}
	if v0 != values1[0] || v1 != values1[1] {
		t.Error("Error: Failed to retrive entries from client cache !")
	}

	// Retrive two entries from the remove the SERVER CACHE
	v2, err := clientDataStore.Retrive(keys1[2])
	if err != nil {
		t.Errorf("Retrive Error: %v path=%v !", err, keys1[2])
	}
	v3, err := clientDataStore.Retrive(keys1[3])
	if err != nil {
		t.Errorf("Retrive Error: %v path=%v !", err, keys1[3])
	}
	if v2 == nil || v3 == nil {
		t.Error("Error: Failed to retrive entries from server cache !")
	}

	// Retrive two entries from the remove the SERVER STORAGE
	v4, err := clientDataStore.Retrive(keys1[4])
	if err != nil {
		t.Errorf("Retrive Error: %v path=%v !", err, keys1[4])
	}
	v5, err := clientDataStore.Retrive(keys1[5])
	if err != nil {
		t.Errorf("Retrive Error: %v path=%v !", err, keys1[5])
	}
	if v4 == nil || v5 == nil {
		t.Error("Error: Failed to retrive entries from server storage !")
	}
	//readonlyClientProxy
}
