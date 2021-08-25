package commutative

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	ccommon "github.com/HPISTechnologies/concurrenturl/common"
)

func TestMeta(t *testing.T) {
	meta := Meta{
		path:      "blcc://eth1.0/accounts/",
		keys:      []string{"0000000000000000000000000000007573657231/", "000000000000000000000000636f696e62617365/"},
		added:     []string{"41eff1c3adfca1ccacced2198241747863dbf800/"},
		removed:   []string{},
		finalized: true,
	}

	data, err := meta.GobEncode()
	if err != nil {
		t.Error("encode err=", err)
		return
	}

	nmeta := Meta{}

	err = nmeta.GobDecode(data)
	if err != nil {
		t.Error("decode err=", err)
		return
	}
	if !reflect.DeepEqual(meta, nmeta) {
		t.Error("Error: Meta Encoding/decoding error, Meta don't match")
	}
	nmeta.Print()
}

func TestBigInt(t *testing.T) {
	avalue := big.NewInt(-252388)

	data, err := avalue.GobEncode()
	if err != nil {
		t.Error("encode err=", err)
		return
	}
	bvalue := big.NewInt(0)
	err = bvalue.GobDecode(data)
	if err != nil {
		t.Error("decode err=", err)
		return
	}
	if !reflect.DeepEqual(avalue, bvalue) {
		t.Error("Error: Bigint Encoding/decoding error, value don't match")
	}
}

func TestBigIntUtil(t *testing.T) {
	avalue := big.NewInt(-252388)
	fmt.Printf("avalue=%v\n", avalue)

	data := ccommon.BigIntEncode(avalue)

	bvalue := ccommon.BigIntDecode(data)

	fmt.Printf("bvalue=%v\n", bvalue)
}

func TestBalance(t *testing.T) {
	balance := Balance{
		value:     big.NewInt(1000000000000000000),
		delta:     big.NewInt(-252388),
		finalized: true,
	}
	balance.Print()
	data, err := balance.GobEncode()
	if err != nil {
		t.Error("encode err=", err)
		return
	}

	nbalance := Balance{}

	err = nbalance.GobDecode(data)
	if err != nil {
		t.Error("decode err=", err)
		return
	}

	nbalance.Print()
	if !reflect.DeepEqual(balance, nbalance) {
		t.Error("Error: Balance Encoding/decoding error, value don't match")
	}
}

func TestInt64(t *testing.T) {
	i64 := Int64{
		value:     int64(1000000000000000000),
		delta:     int64(-252388),
		finalized: true,
	}
	i64.Print()
	data, err := i64.GobEncode()
	if err != nil {
		t.Error("encode err=", err)
		return
	}

	ni64 := Int64{}

	err = ni64.GobDecode(data)
	if err != nil {
		t.Error("decode err=", err)
		return
	}

	ni64.Print()
	if !reflect.DeepEqual(i64.delta, ni64.delta) {
		t.Error("Error: Int64 Encoding/decoding error, value.delta don't match")
	}
	if !reflect.DeepEqual(i64.value, ni64.value) {
		t.Error("Error: Int64 Encoding/decoding error, value.value don't match")
	}
}
