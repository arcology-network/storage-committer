package ccdb

import (
	"fmt"
	"net/http"

	cachedstorage "github.com/arcology/common-lib/cachedstorage"
)

type ReadonlyServer struct {
	addr      string
	dataStore *cachedstorage.DataStore
	encoder   func(interface{}) []byte
	decoder   func([]byte) interface{}
}

func NewReadonlyServer(addr string, encoder func(interface{}) []byte, decoder func([]byte) interface{}, dataStore *cachedstorage.DataStore) *ReadonlyServer {
	return &ReadonlyServer{
		addr:      addr,
		dataStore: dataStore,
		encoder:   encoder,
		decoder:   decoder,
	}
}

func (this *ReadonlyServer) Get(path string) ([]byte, error) {
	val, err := this.dataStore.Retrive(path)
	if err != nil {
		return []byte{}, err
	}
	return this.encoder(val), nil
}

// Get fromt the server connected
func (this *ReadonlyServer) BatchGet(paths []string) ([][]byte, error) {
	bytes := make([][]byte, len(paths))
	for i, v := range paths {
		val, err := this.dataStore.Retrive(v)
		if err != nil {
			fmt.Printf("ReadonlyServer BatchGet err: %v k: %v\n", err, v)
			continue
		}
		bytes[i] = this.encoder(val)
	}
	return bytes, nil
}

func (this *ReadonlyServer) Receive(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		if err := request.ParseForm(); err == nil {
			key := request.FormValue("key")
			if v, _ := this.dataStore.Retrive(key); v != nil {
				writer.Write(this.encoder(v))
			}
		}
	}
}
