package ethrlpbytes

import (
	"encoding/binary"
	"errors"
	"math"
)

type Bytes struct{}

func (Bytes) Encode(input [][]byte) []byte {
	if len(input) == 0 {
		return []byte{0xc0} // Empty list
	}

	body := 0
	for i := 0; i < len(input); i++ {
		if len(input[i]) == 0 {
			body++
			continue
		}

		h, b := precalculate(len(input[i]), input[i][0])
		body += h + b
	}

	var v byte
	if len(input[0]) > 0 {
		v = input[0][0]
	}
	h, b := precalculate(body, v)

	if len(input) == 1 && len(input[0]) <= 1 {
		b = 1
	}

	buffer := make([]byte, h+b)
	offset := len(encodeLengthToVec(buffer, b, 0xc0))
	length := 0

	for _, item := range input {
		length = len(item)
		if length == 1 && item[0] < 0x80 {
			buffer[offset] = item[0]
			offset++
			continue
		}

		offset += len(encodeLengthToVec(buffer[offset:], length, 0x80))
		copy(buffer[offset:], item)
		offset += length
	}
	return buffer
}

func (Bytes) EncodeSingle(item []byte) []byte {
	if len(item) == 0 {
		return []byte{0x80} // Empty list
	}

	header, body := precalculate(len(item), item[0])
	buffer := make([]byte, header+body)

	if len(item) == 1 && item[0] < 0x80 {
		buffer[0] = item[0]
		return buffer[:1] // A single-byte item less than 0x80, encode it as is.
	}

	encodeLengthToVec(buffer, body, 0x80)
	copy(buffer[header:], item)

	return buffer
}

// Decode decodes an RLP-encoded byte slice into a [][]byte slice.
func (Bytes) Decode(encodedBytes []byte) ([][]byte, error) {
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

func precalculate(length int, v byte) (int, int) {
	if length == 1 && v < 0x80 {
		return 1, 0
	}

	if length < 56 {
		return 1, length
	}

	headerBytes := int(math.Ceil(math.Log(float64(length))/math.Log(256))) + 1
	return headerBytes, length
}

func encodeLengthToVec(buffer []byte, length int, offset int) []byte {
	if length < 56 {
		buffer[0] = byte(length + offset)
		return buffer[0:1]
	}

	numBytes := int(math.Ceil(math.Log(float64(length)) / math.Log(256)))
	buffer[0] = byte(numBytes + offset + 55)
	binary.BigEndian.PutUint64(buffer[1:], uint64(length))

	copy(buffer[1:], buffer[9-numBytes:])
	return buffer[:numBytes+1]
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

// decodeLength decodes the length of an item from RLP.
func decodeLength(lengthBytes []byte, offset int) int {
	length := 0
	for _, b := range lengthBytes {
		length = length*256 + int(b)
	}
	return length + offset
}
