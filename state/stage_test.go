// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package state

import (
	"math/big"
	"testing"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/stretchr/testify/assert"
)

func TestStage(t *testing.T) {
	kv, _ := storage.NewMem()
	state, _ := New(polo.Bytes32{}, kv)

	addr := polo.BytesToAddress([]byte("acc1"))

	balance := big.NewInt(10)
	code := []byte{1, 2, 3}

	storage := map[polo.Bytes32]polo.Bytes32{
		polo.BytesToBytes32([]byte("s1")): polo.BytesToBytes32([]byte("v1")),
		polo.BytesToBytes32([]byte("s2")): polo.BytesToBytes32([]byte("v2")),
		polo.BytesToBytes32([]byte("s3")): polo.BytesToBytes32([]byte("v3"))}

	state.SetBalance(addr, balance)
	state.SetCode(addr, code)
	for k, v := range storage {
		state.SetStorage(addr, k, v)
	}

	stage := state.Stage()

	hash, err := stage.Hash()
	assert.Nil(t, err)
	root, err := stage.Commit()
	assert.Nil(t, err)

	assert.Equal(t, hash, root)

	state, _ = New(root, kv)

	assert.Equal(t, balance, state.GetBalance(addr))
	assert.Equal(t, code, state.GetCode(addr))
	assert.Equal(t, polo.Bytes32(crypto.Keccak256Hash(code)), state.GetCodeHash(addr))
	for k, v := range storage {
		assert.Equal(t, v, state.GetStorage(addr, k))
	}
}
