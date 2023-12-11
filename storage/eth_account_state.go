package storage

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	"github.com/arcology-network/concurrenturl/interfaces"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	ethmpt "github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"
)

type Account struct {
	addr string

	ethtypes.StateAccount
	code         []byte
	storageTrie  *ethmpt.Trie // account storage trie
	ethdb        *ethmpt.Database
	diskdbShards [16]ethdb.Database
	err          error

	keyBuffer []string
	valBuffer [][]byte
}

// The diskdbs need to able to handle concurrent accesses themselve
func NewAccount(addr string, diskdbs [16]ethdb.Database, state types.StateAccount) *Account {
	ethdb := ethmpt.NewParallelDatabase(diskdbs, nil)

	trie, err := ethmpt.NewParallel(ethmpt.TrieID(state.Root), ethdb)
	return &Account{
		addr:         addr,
		storageTrie:  trie,
		ethdb:        ethdb,
		diskdbShards: diskdbs,
		StateAccount: state,
		err:          err,
	}
}

func EmptyAccountState() types.StateAccount {
	return ethtypes.StateAccount{
		Nonce:    0,
		Balance:  big.NewInt(0),
		Root:     types.EmptyRootHash,
		CodeHash: codec.Bytes32(crypto.Keccak256Hash(nil)).Encode(),
	}
}

func (this *Account) GetState(key [32]byte) []byte {
	data, _ := this.storageTrie.Get(key[:])
	return data
}

func (this *Account) Trie() *ethmpt.Trie { return this.storageTrie }

func (this *Account) GetCodeHash() [32]byte {
	return codec.Bytes32{}.Decode(this.CodeHash).(codec.Bytes32)
}

func (this *Account) Prove(key [32]byte) ([][]byte, error) {
	var proof proofList
	data, err := this.storageTrie.Get([]byte(key[:]))
	if len(data) > 0 {
		this.storageTrie.Prove([]byte(key[:]), &proof)
	}
	return proof, err
}

func (this *Account) IsProvable(key [32]byte) ([]byte, error) {
	proofs := memorydb.New()
	data, err := this.storageTrie.Get([]byte(key[:]))
	if len(data) > 0 && err == nil {
		if err := this.storageTrie.Prove([]byte(key[:]), proofs); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Failed to find the proof")
	}

	v, err := ethmpt.VerifyProof(this.StateAccount.Root, []byte(key[:]), proofs)
	if err != nil || len(v) == 0 {
		return nil, errors.New("Failed to find the proof")
	}
	return data, nil
}

func (this *Account) DB(key string) ethdb.Database {
	if len(key) == 0 {
		return this.diskdbShards[0]
	}
	return this.diskdbShards[key[0]>>4]
}

func (this *Account) ParseStorageKey(key string) string {
	if k := ccurlcommon.GetPathUnder(key, "/storage/native/"); len(k) > 0 {
		ccurlcommon.GetPathUnder(key, "/storage/native/")
		kstr, err := hexutil.Decode(k)
		if err != nil {
			panic(err)
		}
		return string(kstr)
	}
	return string(this.Hash([]byte(key)))
}

func (this *Account) Has(key string) bool {
	if strings.HasSuffix(key, "/balance") || strings.HasSuffix(key, "/nonce") {
		return true
	}

	if strings.HasSuffix(key, "/code") {
		return len(this.code) > 0
	}

	buffer, _ := this.storageTrie.Get([]byte(this.ParseStorageKey(key)))
	return len(buffer) > 0
}

func (this *Account) Retrive(key string, T any) (interface{}, error) {
	if strings.HasSuffix(key, "/balance") {
		balance, _ := uint256.FromBig(this.StateAccount.Balance)
		v := commutative.NewUnboundedU256()
		v.SetValue(*balance)
		return v, nil
	}

	if strings.HasSuffix(key, "/nonce") {
		v := commutative.NewUnboundedUint64()
		v.SetValue(this.StateAccount.Nonce)
		return v, nil
	}

	if strings.HasSuffix(key, "/code") {
		var err error
		if this.code == nil {
			if this.code, err = this.DB(key).Get(this.CodeHash); err != nil {
				return nil, err
			}
		}
		return noncommutative.NewBytes(this.code), nil
	}

	k := this.ParseStorageKey(key)
	buffer, err := this.storageTrie.Get([]byte(k))
	if len(buffer) == 0 {
		return nil, nil
	}

	if T == nil { // A deletion
		return T, nil
	}
	return T.(interfaces.Type).StorageDecode(buffer), err
}

func (this *Account) UpdateAccountTrie(keys []string, typedVals []interfaces.Type) {
	if pos, _ := common.FindFirstIf(keys, func(k string) bool { return len(k) == ccurlcommon.ETH10_ACCOUNT_FULL_LENGTH+1 }); pos >= 0 {
		common.RemoveAt(&keys, pos)
		common.RemoveAt(&typedVals, pos)
	}

	if pos, _ := common.FindFirstIf(keys, func(k string) bool { return strings.HasSuffix(k, "/nonce") }); pos >= 0 {
		this.Nonce = typedVals[pos].Value().(uint64)
		common.RemoveAt(&keys, pos)
		common.RemoveAt(&typedVals, pos)
	}

	if pos, _ := common.FindFirstIf(keys, func(k string) bool { return strings.HasSuffix(k, "/balance") }); pos >= 0 {
		balance := typedVals[pos].Value().(uint256.Int)
		this.Balance = balance.ToBig()
		common.RemoveAt(&keys, pos)
		common.RemoveAt(&typedVals, pos)
	}

	if pos, _ := common.FindFirstIf(keys, func(k string) bool { return strings.HasSuffix(k, "/code") }); pos >= 0 {
		this.code = typedVals[pos].Value().(codec.Bytes)
		this.StateAccount.CodeHash = this.Hash(this.code)
		if this.DB(keys[pos]).Put(this.CodeHash, this.code) != nil { // Save to DB directly, only for code
			panic("error")
		}
		common.RemoveAt(&keys, pos)
		common.RemoveAt(&typedVals, pos)
	}

	numThd := common.IfThen(len(keys) < 1024, 4, 8)

	k := common.ParallelAppend(keys, numThd, func(i int) string { return this.ParseStorageKey(keys[i]) })
	v := common.ParallelAppend(typedVals, numThd, func(i int) []byte {
		return common.IfThenDo1st(typedVals[i] != nil, func() []byte { return typedVals[i].StorageEncode() }, []byte{})
	})

	// this.storageTrie.ParallelUpdate(codec.Strings(k).ToBytes(), v)

	for i := range k {
		err := this.storageTrie.Update([]byte(k[i]), v[i])
		if err != nil {
			fmt.Printf("*************update err***key=%v**k=%v,%v\n", keys[i], k, err)
		}
		vi, err := this.storageTrie.Get([]byte(k[i]))
		if err != nil {
			fmt.Printf("*********get err*****key=%v****k=%v,%v\n", keys[i], k, err)
		}
		if !bytes.Equal(vi, v[i]) {
			fmt.Printf("*********get err***key=%v*****k=%v,%v\n", keys[i], k, err)
		}
	}

	this.Root = this.storageTrie.Hash()

	this.keyBuffer = k
	this.valBuffer = v

}

func (this *Account) Precommit(keys []string, values []interface{}) {
	this.UpdateAccountTrie(keys, common.Append(values,
		func(v interface{}) interfaces.Type {
			if v.(interfaces.Univalue).Value() != nil {
				return v.(interfaces.Univalue).Value().(interfaces.Type)
			}
			return nil
		}))
}

func (this *Account) Encode() []byte {
	encoded, _ := rlp.EncodeToBytes(&this.StateAccount)
	return encoded
}

func (*Account) Decode(buffer []byte) *Account {
	var acctState types.StateAccount
	rlp.DecodeBytes(buffer, &acctState)
	return &Account{StateAccount: acctState}
}

// Write the DB
func (this *Account) Commit(block uint64) error {
	for i := range this.keyBuffer {
		// err := this.storageTrie.Update([]byte(k[i]), v[i])
		// if err != nil {
		// 	fmt.Printf("*************update err***key=%v**k=%v,%v\n", keys[i], k, err)
		// }
		vi, err := this.storageTrie.Get([]byte(this.keyBuffer[i]))
		if err != nil {
			fmt.Printf("*********get err*********k=%v,%v\n", this.keyBuffer[i], err)
		}
		if !bytes.Equal(vi, this.valBuffer[i]) {
			fmt.Printf("*********get err********k=%v,%v\n", this.keyBuffer[i], err)
		}
	}

	root, nodes, err := this.storageTrie.Commit(false) // Finalized the trie
	if err != nil {
		return err
	}

	if root != this.Root {
		return errors.New("Root mismatched")
	}

	// Initialize an empty node set, even there is no change
	nodes = common.IfThen(nodes == nil, trienode.NewNodeSet(types.EmptyRootHash), nodes)

	if err := this.ethdb.Update(root, types.EmptyRootHash, block, trienode.NewWithNodeSet(nodes), nil); err != nil { // Move to DB dirty node set
		return err
	}

	if err := this.ethdb.Commit(root, false); err != nil { // Write to DB
		this.ethdb.Commit(root, false)
		return err
	}

	this.storageTrie, err = ethmpt.NewParallel(ethmpt.TrieID(root), this.ethdb)

	return err
	// return this.ethdb.Commit(root, false) // Write to DB
}

func (this *Account) Hash(key []byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(key))
	sum := hasher.Sum(nil)
	return sum
}

func (this *Account) Print() {
	fmt.Println("addr: ", this.addr)
	fmt.Println("StateAccount: ", this.StateAccount)
	fmt.Println("code: ", this.code)
	fmt.Println("storageTrie: ", this.storageTrie)
	fmt.Println("ethdb: ", this.ethdb)
	fmt.Println("diskdbShards: ", this.diskdbShards)
	fmt.Println("err: ", this.err)
}
