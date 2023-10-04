package merklepatriciatrie

// import (
// 	"bytes"
// 	"encoding/binary"
// 	"errors"
// 	"math/big"
// )

// // Encode encodes a [][]byte slice into an RLP-encoded byte slice.
// func Encode(input [][]byte) []byte {
// 	var encodedBytes []byte

// 	for _, item := range input {
// 		encodedBytes = append(encodedBytes, encodeItem(item)...)
// 	}

// 	return append(encodeLength(len(encodedBytes), 0xc0), encodedBytes...)
// }

// // Decode decodes an RLP-encoded byte slice into a [][]byte slice.
// func Decode(encodedBytes []byte) ([][]byte, error) {
// 	var output [][]byte

// 	reader := bytes.NewReader(encodedBytes)

// 	for reader.Len() > 0 {
// 		item, err := decodeItem(reader)
// 		if err != nil {
// 			return nil, err
// 		}
// 		output = append(output, item)
// 	}

// 	return output, nil
// }

// // encodeItem encodes a single []byte item.
// func encodeItem(item []byte) []byte {
// 	if len(item) == 1 && item[0] < 0x80 {
// 		return item // A single-byte item less than 0x80, encode it as is.
// 	}

// 	return append(encodeLength(len(item), 0x80), item...)
// }

// // decodeItem decodes a single []byte item.
// func decodeItem(reader *bytes.Reader) ([]byte, error) {
// 	b, err := reader.ReadByte()
// 	if err != nil {
// 		return nil, err
// 	}

// 	if b < 0x80 {
// 		return []byte{b}, nil // Single byte item
// 	} else if b < 0xB8 {
// 		length := int(b) - 0x80
// 		data := make([]byte, length)
// 		_, err := reader.Read(data)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return data, nil
// 	} else if b < 0xC0 {
// 		lengthBytes := make([]byte, int(b)-0xB7)
// 		_, err := reader.Read(lengthBytes)
// 		if err != nil {
// 			return nil, err
// 		}
// 		length := decodeLength(lengthBytes, 0x80)
// 		data := make([]byte, length)
// 		_, err = reader.Read(data)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return data, nil
// 	} else if b < 0xF8 {
// 		listLength := int(b) - 0xC0
// 		var data []byte
// 		for listLength > 0 {
// 			item, err := decodeItem(reader)
// 			if err != nil {
// 				return nil, err
// 			}
// 			data = append(data, item...)
// 			listLength--
// 		}
// 		return data, nil
// 	} else {
// 		return nil, errors.New("invalid RLP encoding")
// 	}
// }

// // encodeLength encodes the length of an item for RLP.
// func encodeLength(length int, offset int) []byte {
// 	if length < 56 {
// 		return []byte{byte(length + offset)}
// 	}

// 	// Encode the length as a binary string
// 	lengthBytes := make([]byte, 8)
// 	bigLength := big.NewInt(int64(length))
// 	bigLength.FillBytes(lengthBytes)
// 	var i int
// 	for i < len(lengthBytes) && lengthBytes[i] == 0 {
// 		i++
// 	}
// 	lengthBytes = lengthBytes[i:]

// 	// Add the offset to the length code
// 	buffer := make([]byte, 4)
// 	binary.LittleEndian.PutUint32(buffer, uint32(len(lengthBytes)+offset))
// 	buffer = append(buffer, 0x38)

// 	return append(buffer, lengthBytes...)
// }

// // decodeLength decodes the length of an item from RLP.
// func decodeLength(lengthBytes []byte, offset int) int {
// 	length := 0
// 	for _, b := range lengthBytes {
// 		length = length*256 + int(b)
// 	}
// 	return length + offset
// }
