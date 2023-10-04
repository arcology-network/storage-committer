package merklepatriciatrie

import (
	"encoding/binary"
	"errors"
	"math/big"
)

func Encode(input [][]byte) []byte {
	if len(input) == 0 {
		return []byte{0xc0} // Empty list
	}

	total := 0
	for i := 0; i < len(input); i++ {
		total += len(input[i])
	}

	encodedBytes := make([]byte, 0, total+len(input)+4+8)
	for _, item := range input {
		if len(item) == 1 && item[0] < 0x80 {
			encodedBytes = append(encodedBytes, item[0]) // A single-byte item less than 0x80, encode it as is.
			continue
		}

		lengthBytes := encodeLength(len(item), 0x80)
		encodedBytes = append(encodedBytes, lengthBytes...)
		encodedBytes = append(encodedBytes, item...)
	}
	return append(encodeLength(len(encodedBytes), 0xc0), encodedBytes...)
}

// Decode decodes an RLP-encoded byte slice into a [][]byte slice.
func Decode(encodedBytes []byte) ([][]byte, error) {
	output := make([][]byte, 0, 32)

	for len(encodedBytes) > 0 {
		item, bytesRead, err := decodeItem(encodedBytes)
		if err != nil {
			return nil, err
		}
		output = append(output, item)
		encodedBytes = encodedBytes[bytesRead:]
	}

	return output, nil
}

// decodeItem decodes a single RLP-encoded item.
func decodeItem(encodedBytes []byte) ([]byte, int, error) {
	if len(encodedBytes) == 0 {
		return nil, 0, errors.New("empty RLP item")
	}

	b := encodedBytes[0]
	if b < 0x80 {
		return []byte{b}, 1, nil // Single byte item
	}

	if b < 0xB8 {
		length := int(b) - 0x80
		if length > len(encodedBytes)-1 {
			return nil, 0, errors.New("insufficient data for RLP item")
		}
		return encodedBytes[1 : 1+length], 1 + length, nil
	}

	if b < 0xC0 {
		lengthLen := int(b) - 0xB7
		if lengthLen > len(encodedBytes)-1 {
			return nil, 0, errors.New("insufficient data for RLP item")
		}

		lengthBytes := encodedBytes[1 : 1+lengthLen]
		length := decodeLength(lengthBytes, 0x80)
		if length+len(lengthBytes)+1 > len(encodedBytes) {
			return nil, 0, errors.New("insufficient data for RLP item")
		}

		data := encodedBytes[1+lengthLen : 1+lengthLen+length]
		return data, 1 + lengthLen + length, nil
	}

	if b < 0xF8 {
		listLength := int(b) - 0xC0
		offset := 1
		data := make([]byte, 0, len(encodedBytes))

		for listLength > 0 {
			if offset >= len(encodedBytes) {
				return nil, 0, errors.New("insufficient data for RLP item")
			}

			item, bytesRead, err := decodeItem(encodedBytes[offset:])
			if err != nil {
				return nil, 0, err
			}

			data = append(data, item...)
			offset += bytesRead
			listLength--
		}
		return data, offset, nil
	}

	return nil, 0, errors.New("invalid RLP encoding")
}

// encodeLength encodes the length of an item for RLP.
func encodeLength(length int, offset int) []byte {
	if length < 56 {
		return []byte{byte(length + offset)}
	}

	// Add the offset to the length code
	buffer := make([]byte, 13)
	binary.LittleEndian.PutUint32(buffer[:4], uint32(8+offset))
	buffer[4] = 0x38

	// Encode the length as a binary string
	bigLength := big.NewInt(int64(length))

	lengthBytes := buffer[5:] //make([]byte, 8)
	bigLength.FillBytes(lengthBytes)

	var i int
	for i < len(lengthBytes) && lengthBytes[i] == 0 {
		i++
	}
	lengthBytes = lengthBytes[i:] // Remove the leading 0s

	copy(buffer[5:], lengthBytes)
	buffer = buffer[:5+len(lengthBytes)]

	return buffer //[:5+len(lengthBytes)] //append(buffer[:5 + len(lengthBytes)], lengthBytes...)
}

// decodeLength decodes the length of an item from RLP.
func decodeLength(lengthBytes []byte, offset int) int {
	length := 0
	for _, b := range lengthBytes {
		length = length*256 + int(b)
	}
	return length + offset
}
