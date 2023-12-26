package ccurltest

import (
	"errors"
	"reflect"

	"github.com/arcology-network/common-lib/common"
	orderedset "github.com/arcology-network/common-lib/container/set"
	ccurl "github.com/arcology-network/concurrenturl"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	"github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/interfaces"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
)

func Create_Ctrn_0(account string, store interfaces.Datastore) ([]byte, []interfaces.Univalue, error) {
	url := ccurl.NewConcurrentUrl(store)
	writeCache := url.WriteCache()
	path := commutative.NewPath() // create a path
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/", path); err != nil {
		return []byte{}, nil, err
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-00", noncommutative.NewString("tx0-elem-00")); err != nil { /* The first Element */
		return []byte{}, nil, err
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-01", noncommutative.NewString("tx0-elem-01")); err != nil { /* The second Element */
		return []byte{}, nil, err
	}

	rawTrans := url.WriteCache().Export(indexer.Sorter)
	transitions := indexer.Univalues(common.Clone(rawTrans)).To(indexer.ITCTransition{})
	return indexer.Univalues(transitions).Encode(), transitions, nil
}

func ParallelInsert_Ctrn_0(account string, store interfaces.Datastore) ([]byte, error) {
	url := ccurl.NewConcurrentUrl(store)
	writeCache := url.WriteCache()
	path := commutative.NewPath() // create a path
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/", path); err != nil {
		return []byte{}, err
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-00", noncommutative.NewString("tx0-elem-00")); err != nil { /* The first Element */
		return []byte{}, err
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-01", noncommutative.NewString("tx0-elem-01")); err != nil { /* The second Element */
		return []byte{}, err
	}

	transitions := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.ITCTransition{})
	return indexer.Univalues(transitions).Encode(), nil
}

func Create_Ctrn_1(account string, store interfaces.Datastore) ([]byte, error) {
	url := ccurl.NewConcurrentUrl(store)
	writeCache := url.WriteCache()
	path := commutative.NewPath() // create a path
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/", path); err != nil {
		return []byte{}, err
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/elem-00", noncommutative.NewString("tx1-elem-00")); err != nil { /* The first Element */
		return []byte{}, err
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/elem-01", noncommutative.NewString("tx1-elem-00")); err != nil { /* The second Element */
		return []byte{}, err
	}

	transitions := indexer.Univalues(common.Clone(url.WriteCache().Export(indexer.Sorter))).To(indexer.ITCTransition{})
	return indexer.Univalues(transitions).Encode(), nil
}

func CheckPaths(account string, url *ccurl.ConcurrentUrl) error {
	v, _ := url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-00", new(noncommutative.String))
	if v.(string) != "tx0-elem-00" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-01", new(noncommutative.String))
	if v.(string) != "tx0-elem-01" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/elem-00", new(noncommutative.String))
	if v.(string) != "tx1-elem-00" {
		return errors.New("Error: Not match")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/elem-01", new(noncommutative.String))
	if v.(string) != "tx1-elem-00" {
		return errors.New("Error: Not match")
	}

	//Read the path
	v, _ = url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/", new(commutative.Path))
	keys := v.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Path don't match !")
	}

	// Read the path again
	v, _ = url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/", new(commutative.Path))
	keys = v.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Path don't match !")
	}

	v, _ = url.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/", new(commutative.Path))
	keys = v.(*orderedset.OrderedSet).Keys()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Path don't match !")
	}
	return nil
}
