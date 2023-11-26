package ccdb

import (
	"math/big"
	"strings"

	"github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	"github.com/arcology-network/concurrenturl/interfaces"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	"github.com/arcology-network/evm/core/types"
	"github.com/arcology-network/evm/ethdb"
	"github.com/arcology-network/evm/rlp"
	ethmpt "github.com/arcology-network/evm/trie"
	"github.com/arcology-network/evm/trie/trienode"
	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"
)

type Account struct {
	addr string
	types.StateAccount
	code         []byte
	storageTrie  *ethmpt.Trie // account storage trie
	ethdb        *ethmpt.Database
	diskdbShards [16]ethdb.Database
}

// The diskdbs need to able to handle concurrent accesses themselve
func NewAccount(addr string, diskdbs [16]ethdb.Database, state types.StateAccount) *Account {
	ethdb := ethmpt.NewParallelDatabase(diskdbs, nil)

	trie, _ := ethmpt.NewParallel(ethmpt.TrieID(state.Root), ethdb)
	return &Account{
		addr:         addr,
		storageTrie:  trie,
		ethdb:        ethdb,
		diskdbShards: diskdbs,
		StateAccount: state,
	}
}

func EmptyAccountState() types.StateAccount {
	return types.StateAccount{
		Nonce:    0,
		Balance:  big.NewInt(0),
		Root:     types.EmptyRootHash,
		CodeHash: types.EmptyCodeHash[:],
	}
}

func (this *Account) SelectDB(key string) ethdb.Database {
	if len(key) == 0 {
		return this.diskdbShards[0]
	}
	return this.diskdbShards[key[0]>>4]
}

func (this *Account) storageKey(key string) string {
	if k := ccurlcommon.UnderNative(key); len(k) > 0 {
		return k
	}
	return string(this.Hash([]byte(key)))
}

func (this *Account) IfExists(key string) bool {
	if strings.HasSuffix(key, "/balance") || strings.HasSuffix(key, "/nonce") {
		return true
	}

	if strings.HasSuffix(key, "/code") {
		return len(this.code) > 0
	}

	buffer, _ := this.storageTrie.Get([]byte(this.storageKey(key)))
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
			if this.code, err = this.SelectDB(key).Get(this.CodeHash); err != nil {
				return nil, err
			}
		}
		return noncommutative.NewBytes(this.code), nil
	}

	key = this.storageKey(key)
	buffer, err := this.storageTrie.Get([]byte(key))
	if len(buffer) == 0 {
		return nil, nil
	}

	if T == nil { // A deletion
		return T, nil
	}
	return T.(interfaces.Type).StorageDecode(buffer), err
}

func (this *Account) UpdateAccountTrie(keys []string, typedVals []interfaces.Type) {
	if pos, _ := common.FindFirstIf(keys, func(v string) bool { return strings.HasSuffix(v, "/nonce") }); pos >= 0 {
		this.Nonce = typedVals[pos].Value().(uint64)
		common.RemoveAt(&keys, pos)
		common.RemoveAt(&typedVals, pos)
	}

	if pos, _ := common.FindFirstIf(keys, func(key string) bool { return strings.HasSuffix(key, "/balance") }); pos >= 0 {
		balance := typedVals[pos].Value().(uint256.Int)
		this.Balance = balance.ToBig()
		common.RemoveAt(&keys, pos)
		common.RemoveAt(&typedVals, pos)
	}

	if pos, _ := common.FindFirstIf(keys, func(key string) bool { return strings.HasSuffix(key, "/code") }); pos >= 0 {
		this.code = typedVals[pos].Value().(codec.Bytes)
		this.StateAccount.CodeHash = this.Hash(this.code)
		if this.SelectDB(keys[pos]).Put(this.CodeHash, this.code) != nil { // Save to DB directly, only for code
			panic("error")
		}
		common.RemoveAt(&keys, pos)
		common.RemoveAt(&typedVals, pos)
	}

	k := common.ParallelAppend(keys, func(i int) []byte { return []byte(this.storageKey(keys[i])) })
	v := common.ParallelAppend(typedVals, func(i int) []byte {
		return common.IfThenDo1st(typedVals[i] != nil, func() []byte { return typedVals[i].StorageEncode() }, []byte{})
	})

	// common.Foreach(keys, func(_ *string, i int) { this.storageTrie.Update(k[i], v[i]) }) // Sequential

	this.storageTrie.ParallelUpdate(k, v)
	this.Root = this.storageTrie.Hash()
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

// Write the DB
func (this *Account) Commit() error {
	root, nodes := this.storageTrie.Commit(false)                                                        // Finalized the trie
	if err := this.ethdb.Update(root, types.EmptyRootHash, trienode.NewWithNodeSet(nodes)); err != nil { // Move to DB dirty node set
		return err
	}
	return this.ethdb.Commit(root, false) // Write to DB
}

func (*Account) Decode(buffer []byte) *Account {
	var acctState types.StateAccount
	rlp.DecodeBytes(buffer, acctState)
	return &Account{StateAccount: acctState}
}

func (this *Account) Hash(key []byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(key))
	sum := hasher.Sum(nil)
	return sum
}
