package merklepatriciatrie

import (
	"fmt"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/merkle"
	"github.com/arcology-network/evm/rlp"
	"github.com/stretchr/testify/require"
	// "github.com/HPISTechnologies/evm/rlp"
)

func TestEncodeDecode(t *testing.T) {
	data := [][]byte{
		{uint8(0)},
		{uint8(56)},
		{uint8(255)},
		[]byte(""),
		[]byte("Hello"),
		[]byte("World,World,World"),
	}

	v, err := rlp.EncodeToBytes(data)
	if err != nil {
		t.Error(v, err)
	}

	encoded := Encode(data)
	fmt.Printf("Encoded: %x\n", encoded)

	require.Equal(t, v, encoded)

	// decoded, err := Decode(encoded)
	// if err != nil {
	// 	fmt.Println("Error decoding RLP:", err)
	// 	return
	// }

	// fmt.Printf("Decoded: %v\n", decoded)
	// require.Equal(t, data, decoded)

	// var decodedv []interface{}
	// if err := rlp.DecodeBytes(encoded, &decodedv); err != nil {
	// 	log.Fatal("Error decoding data:", err)
	// }
	// require.Equal(t, data, decodedv)

}

func TestEncodeDecodePerformance(t *testing.T) {
	data := make([][]byte, 1000000)
	// data := make([][]byte, len(keys))
	// t0 := time.Now()
	for i := 0; i < len(data); i++ {
		data[i] = merkle.Sha256{}.Hash([]byte(fmt.Sprint(i)))
		// data[i] = merkle.Sha256{}.Hash([]byte(fmt.Sprint(i)))
	}

	t0 := time.Now()
	for i := 0; i < len(data); i++ {
		rlp.EncodeToBytes(data[i])
	}
	fmt.Println("RLP universal loop: ", time.Since(t0))

	t0 = time.Now()
	rlp.EncodeToBytes(data)
	fmt.Println("RLP universal array: ", time.Since(t0))

	t0 = time.Now()
	// for i := 0; i < len(data); i++ {
	Encode(data)
	// }
	fmt.Println("Speicallized array: ", time.Since(t0))
}
