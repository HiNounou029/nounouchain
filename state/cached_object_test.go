// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package state

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/HiNounou029/nounouchain/trie"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)

func TestCachedObject(t *testing.T) {
	kv, _ := storage.NewMem()

	stgTrie, _ := trie.NewSecure(polo.Bytes32{}, kv, 0)
	storages := []struct {
		k polo.Bytes32
		v rlp.RawValue
	}{
		{polo.BytesToBytes32([]byte("key1")), []byte("value1")},
		{polo.BytesToBytes32([]byte("key2")), []byte("value2")},
		{polo.BytesToBytes32([]byte("key3")), []byte("value3")},
		{polo.BytesToBytes32([]byte("key4")), []byte("value4")},
	}

	for _, s := range storages {
		saveStorage(stgTrie, s.k, s.v)
	}

	storageRoot, _ := stgTrie.Commit()

	code := make([]byte, 100)
	rand.Read(code)

	codeHash := crypto.Keccak256(code)
	kv.Put(codeHash, code)

	account := Account{
		Balance:     &big.Int{},
		CodeHash:    codeHash,
		StorageRoot: storageRoot[:],
	}

	obj := newCachedObject(kv, &account)

	assert.Equal(t,
		M(obj.GetCode()),
		[]interface{}{code, nil})

	for _, s := range storages {
		assert.Equal(t,
			M(obj.GetStorage(s.k)),
			[]interface{}{s.v, nil})
	}
}
