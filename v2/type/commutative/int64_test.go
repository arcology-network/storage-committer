package commutative

// func TestCodecCommutative(t *testing.T) {
// 	/* Commutative Int64 Test */
// 	inInt64 := NewInt64(12345, 0)
// 	int64Bytes := inInt64.(*Int64).Encode()
// 	outInt64 := (&Int64{}).Decode(int64Bytes).(*Int64)
// 	if !reflect.DeepEqual(inInt64, outInt64) {
// 		t.Error("Error: Int64 Encoding/decoding error, numbers don't match")
// 	}

// 	/* Commutative Bigint Test */
// 	inBig := NewBalance(big.NewInt(789456), big.NewInt(0))
// 	bigBytes := inBig.(*Balance).Encode()
// 	outBig := (&Balance{}).Decode(bigBytes).(*Balance)
// 	if !reflect.DeepEqual(inBig, outBig) {
// 		t.Error("Error: Bigint Encoding/decoding error, numbers don't match")
// 	}
// }
