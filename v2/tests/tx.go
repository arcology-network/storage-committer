package ccurltest

import (
	"errors"
	"reflect"

	cachedstorage "github.com/arcology-network/common-lib/cachedstorage"
	ccurl "github.com/arcology-network/concurrenturl/v2"
	ccurlcommon "github.com/arcology-network/concurrenturl/v2/common"
	commutative "github.com/arcology-network/concurrenturl/v2/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/v2/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/v2/univalue"
)

func SimulatedTx0(account string, store *cachedstorage.DataStore) ([]byte, error) {
	url := ccurl.NewConcurrentUrl(store)
	path := commutative.NewPath() // create a path
	if err := url.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/", path); err != nil {
		return []byte{}, err
	}

	if err := url.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-00", noncommutative.NewString("tx0-elem-00")); err != nil { /* The first Element */
		return []byte{}, err
	}

	if err := url.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-01", noncommutative.NewString("tx0-elem-01")); err != nil { /* The second Element */
		return []byte{}, err
	}

	_, transitions := url.Export3(ccurlcommon.Sorter)
	return univalue.Univalues(transitions).Encode(), nil
}

func SimulatedTx1(account string, store *cachedstorage.DataStore) ([]byte, error) {
	url := ccurl.NewConcurrentUrl(store)
	path := commutative.NewPath() // create a path
	if err := url.Write(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/", path); err != nil {
		return []byte{}, err
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/elem-00", noncommutative.NewString("tx1-elem-00")); err != nil { /* The first Element */
		return []byte{}, err
	}

	if err := url.Write(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/elem-01", noncommutative.NewString("tx1-elem-00")); err != nil { /* The second Element */
		return []byte{}, err
	}

	_, transitions := url.Export3(ccurlcommon.Sorter)
	return univalue.Univalues(transitions).Encode(), nil
}

func CheckPaths(account string, url *ccurl.ConcurrentUrl) error {
	v, _ := url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-00")
	if *v.(*noncommutative.String) != "tx0-elem-00" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-01")
	if *v.(*noncommutative.String) != "tx0-elem-01" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/elem-00")
	if *v.(*noncommutative.String) != "tx1-elem-00" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/elem-01")
	if *v.(*noncommutative.String) != "tx1-elem-00" {
		return errors.New("Error: Not match")
	}

	//Read the path
	v, _ = url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/")
	keys := v.([]string)
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Path don't match !")
	}

	// Read the path again
	v, _ = url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/")
	if !reflect.DeepEqual(v.([]string), []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Path don't match !")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/")
	if !reflect.DeepEqual(v.([]string), []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Path don't match !")
	}
	return nil
}
