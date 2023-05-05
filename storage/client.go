package ccdb

import (
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/arcology-network/common-lib/cachedstorage"
	datacompression "github.com/arcology-network/common-lib/datacompression"
)

type ReadonlyClient struct {
	addr         string
	path         string
	uncompressor *datacompression.CompressionLut
	localStore   *cachedstorage.DataStore
}

func NewReadonlyClient(addr string, path string, lut *datacompression.CompressionLut, args ...interface{}) *ReadonlyClient {
	readonlyClient := &ReadonlyClient{
		addr:         addr,
		path:         path,
		uncompressor: lut,
	}

	if len(args) > 0 && args[0] != nil {
		readonlyClient.localStore = args[0].(*cachedstorage.DataStore)
	}
	return readonlyClient
}

// Get from the server connected
func (this *ReadonlyClient) Get(key string) ([]byte, error) {
	if this.localStore != nil {
		v, err := this.localStore.Retrive(key)
		if err != nil {
			return []byte{}, err
		}
		bytes := this.localStore.Encoder()(v)

		return bytes, err
	} else {
		base, err := url.Parse(this.addr)
		if err != nil {
			return nil, errors.New("Error: The website is unreachable !")
		}

		base.Path = this.path
		params := url.Values{}
		params.Add("key", key)
		base.RawQuery = params.Encode()

		resp, err := http.Get(base.String())
		if err != nil {
			return nil, err
		}

		bytes, err := io.ReadAll(resp.Body)
		if err == nil {
			return bytes, nil
		}
		return nil, err
	}
}

// Get fromt the server connected
func (this *ReadonlyClient) BatchGet(keys []string) ([][]byte, error) {
	if this.localStore != nil {
		results := make([][]byte, len(keys))
		for i := 0; i < len(keys); i++ {
			results[i], _ = this.Get(keys[i])
		}
		return results, nil

	} else {
		// Get from the server
		return nil, nil
	}
}

// Ready only, do nothing
func (*ReadonlyClient) Set(path string, v []byte) error           { return nil }
func (*ReadonlyClient) BatchSet(paths []string, v [][]byte) error { return nil }
func (*ReadonlyClient) Query(pattern string, condition func(string, string) bool) ([]string, [][]byte, error) {
	return []string{}, [][]byte{}, nil
}
