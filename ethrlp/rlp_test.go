package ethrlpbytes

import (
	"fmt"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/merkle"
	"github.com/arcology-network/evm/rlp"
	"github.com/stretchr/testify/require"
	// "github.com/HPISTechnologies/evm/rlp"
)

func TestEncodeDecodeSingle(t *testing.T) {
	emptyv := []byte{}
	v := []byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello")
	longv := []byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello")

	t.Run("Original RLP", func(t *testing.T) {
		expected, err := rlp.EncodeToBytes(v)
		if err != nil {
			t.Error(expected, err)
		}

		var decoded []byte
		if err := rlp.DecodeBytes(expected, &decoded); err != nil {
			t.Error(expected, err)
		}
		require.Equal(t, v, decoded)
	})

	t.Run("Simplified RLP 0 ~ 255", func(t *testing.T) {
		for i := 0; i < 256; i++ {
			encoded := Bytes{}.EncodeSingle([]byte{uint8(i)})

			var decoded []byte
			if err := rlp.DecodeBytes(encoded, &decoded); err != nil {
				t.Error(encoded, err)
			}
			require.Equal(t, []byte{uint8(i)}, decoded)
		}
	})

	t.Run("Simplified empty", func(t *testing.T) {
		encoded := Bytes{}.EncodeSingle(emptyv)

		var decoded []byte
		if err := rlp.DecodeBytes(encoded, &decoded); err != nil {
			t.Error(encoded, err)
		}
		require.Equal(t, emptyv, decoded)
	})

	t.Run("Simplified RLP", func(t *testing.T) {
		encoded := Bytes{}.EncodeSingle(v)

		var decoded []byte
		if err := rlp.DecodeBytes(encoded, &decoded); err != nil {
			t.Error(encoded, err)
		}
		require.Equal(t, v, decoded)
	})

	t.Run("Simplified RLP long", func(t *testing.T) {
		encoded := Bytes{}.EncodeSingle(longv)

		var decoded []byte
		if err := rlp.DecodeBytes(encoded, &decoded); err != nil {
			t.Error(encoded, err)
		}
		require.Equal(t, longv, decoded)
	})
}

func TestEncodeDecode(t *testing.T) {
	t.Run("Single empty string", func(t *testing.T) {
		data := [][]byte{[]byte("")}
		expected, _ := rlp.EncodeToBytes(data)
		require.Equal(t, expected, Bytes{}.Encode(data))
	})

	t.Run("Multiple single long strings", func(t *testing.T) {
		data := [][]byte{
			[]byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello")}
		expected, _ := rlp.EncodeToBytes(data)
		require.Equal(t, expected, Bytes{}.Encode(data))
	})

	t.Run("Multiple multiple long strings", func(t *testing.T) {
		data := [][]byte{
			[]byte(""),
			[]byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello"),
			[]byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello"),
			[]byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello"),
			[]byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello"),
			[]byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello"),
			[]byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello"),
			[]byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello"),
			[]byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello"),
			[]byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello"),
			[]byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello"),
			[]byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello"),
			[]byte(""),
		}
		expected, _ := rlp.EncodeToBytes(data)
		require.Equal(t, expected, Bytes{}.Encode(data))
	})

	t.Run("Multiple mixed long strings", func(t *testing.T) {
		data := [][]byte{{uint8(0)}, {uint8(56)}, {uint8(255)},
			{uint8(255)}, []byte("Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello,Hello"), []byte("hhhh"), []byte("255"), []byte("")}
		expected, _ := rlp.EncodeToBytes(data)
		require.Equal(t, expected, Bytes{}.Encode(data))
	})

	t.Run("Multiple empty strings", func(t *testing.T) {
		data := [][]byte{[]byte(""), []byte("")}
		expected, _ := rlp.EncodeToBytes(data)
		require.Equal(t, expected, Bytes{}.Encode(data))
	})

	t.Run("Multiple non empty strings", func(t *testing.T) {
		data := [][]byte{{uint8(0)}, {uint8(56)}, {uint8(255)}, []byte(""), []byte("Hello"), []byte("hhhh"), []byte("255"), []byte("")}
		expected, _ := rlp.EncodeToBytes(data)
		require.Equal(t, expected, Bytes{}.Encode(data))
	})

	t.Run("Multiple mixed strings", func(t *testing.T) {
		data := [][]byte{[]byte(""), {uint8(0)}, {uint8(56)}, {uint8(255)}, []byte(""), []byte("Hello"), []byte("hhhh"), []byte("255"), []byte("")}
		expected, _ := rlp.EncodeToBytes(data)
		require.Equal(t, expected, Bytes{}.Encode(data))
	})
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
	for i := 0; i < len(data); i++ {
		Bytes{}.EncodeSingle(data[i])
	}
	fmt.Println("EncodeSingle loop: ", time.Since(t0))

	t0 = time.Now()
	// for i := 0; i < len(data); i++ {
	Bytes{}.Encode(data)
	// }
	fmt.Println("Speicallized array: ", time.Since(t0))
}
