package common

import (
	"bytes"
	"crypto/sha256"

	orderedmap "github.com/elliotchance/orderedmap"
)

type Checksum struct {
	checksums *orderedmap.OrderedMap
}

func (this *Checksum) NewChecksum() *Checksum {
	return &Checksum{
		checksums: orderedmap.NewOrderedMap(),
	}
}

func (this *Checksum) Check(keys []string, datasource *DataStore) []string {
	corrupted := []string{}
	for _, key := range keys {
		encoded := datasource.Retrive(key).(TypeInterface).Encode()

		current := sha256.Sum256(encoded)
		if v, ok := this.checksums.Get(current[:]); ok {
			previous := v.([32]byte)
			if bytes.Equal(previous[:], current[:]) {
				corrupted = append(corrupted, key)
			}
		}
	}
	return corrupted
}

func (this *Checksum) Update(keys []string, states []interface{}) {
	for i, state := range states {
		encoded := state.(TypeInterface).Encode()
		current := sha256.Sum256(encoded)
		this.checksums.Set(keys[i], current[:])
	}
}
