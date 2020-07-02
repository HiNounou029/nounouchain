// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package state

import (
	"math/big"
	"testing"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/HiNounou029/nounouchain/trie"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)

func M(a ...interface{}) []interface{} {
	return a
}

func TestAccount(t *testing.T) {
	assert.True(t, emptyAccount().IsEmpty())

	acc := emptyAccount()
	acc.Balance = big.NewInt(1)
	assert.False(t, acc.IsEmpty())
	acc = emptyAccount()
	acc.CodeHash = []byte{1}
	assert.False(t, acc.IsEmpty())

	acc = emptyAccount()
	assert.True(t, acc.IsEmpty())

	acc = emptyAccount()
	acc.StorageRoot = []byte{1}
	assert.True(t, acc.IsEmpty())
}

func newTrie() *trie.SecureTrie {
	kv, _ := storage.NewMem()
	trie, _ := trie.NewSecure(polo.Bytes32{}, kv, 0)
	return trie
}
func TestTrie(t *testing.T) {
	trie := newTrie()

	addr := polo.BytesToAddress([]byte("account1"))
	assert.Equal(t,
		M(loadAccount(trie, addr)),
		[]interface{}{emptyAccount(), nil},
		"should load an empty account")

	acc1 := Account{
		big.NewInt(1),
		0,
		[]byte("master"),
		[]byte("code hash"),
		[]byte("storage root"),
	}
	saveAccount(trie, addr, &acc1)
	assert.Equal(t,
		M(loadAccount(trie, addr)),
		[]interface{}{&acc1, nil})

	saveAccount(trie, addr, emptyAccount())
	assert.Equal(t,
		M(trie.TryGet(addr[:])),
		[]interface{}{[]byte(nil), nil},
		"empty account should be deleted")
}

func TestStorageTrie(t *testing.T) {
	trie := newTrie()

	key := polo.BytesToBytes32([]byte("key"))
	assert.Equal(t,
		M(loadStorage(trie, key)),
		[]interface{}{rlp.RawValue(nil), nil})

	value := rlp.RawValue("value")
	saveStorage(trie, key, value)
	assert.Equal(t,
		M(loadStorage(trie, key)),
		[]interface{}{value, nil})

	saveStorage(trie, key, nil)
	assert.Equal(t,
		M(trie.TryGet(key[:])),
		[]interface{}{[]byte(nil), nil},
		"empty storage value should be deleted")
}
