// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package state

import (
	"math/big"
	"testing"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)

func TestStateReadWrite(t *testing.T) {
	kv, _ := storage.NewMem()
	state, _ := New(polo.Bytes32{}, kv)

	addr := polo.BytesToAddress([]byte("account1"))
	storageKey := polo.BytesToBytes32([]byte("storageKey"))

	assert.False(t, state.Exists(addr))
	assert.Equal(t, state.GetBalance(addr), &big.Int{})
	assert.Equal(t, state.GetCode(addr), []byte(nil))
	assert.Equal(t, state.GetCodeHash(addr), polo.Bytes32{})
	assert.Equal(t, state.GetStorage(addr, storageKey), polo.Bytes32{})

	// make account not empty
	state.SetBalance(addr, big.NewInt(1))
	assert.Equal(t, state.GetBalance(addr), big.NewInt(1))

	state.SetMaster(addr, polo.BytesToAddress([]byte("master")))
	assert.Equal(t, polo.BytesToAddress([]byte("master")), state.GetMaster(addr))

	state.SetCode(addr, []byte("code"))
	assert.Equal(t, state.GetCode(addr), []byte("code"))
	assert.Equal(t, state.GetCodeHash(addr), polo.Bytes32(crypto.Keccak256Hash([]byte("code"))))

	assert.Equal(t, state.GetStorage(addr, storageKey), polo.Bytes32{})
	state.SetStorage(addr, storageKey, polo.BytesToBytes32([]byte("storageValue")))
	assert.Equal(t, state.GetStorage(addr, storageKey), polo.BytesToBytes32([]byte("storageValue")))

	assert.True(t, state.Exists(addr))

	// delete account
	state.Delete(addr)
	assert.False(t, state.Exists(addr))
	assert.Equal(t, state.GetBalance(addr), &big.Int{})
	assert.Equal(t, state.GetMaster(addr), polo.Address{})
	assert.Equal(t, state.GetCode(addr), []byte(nil))
	assert.Equal(t, state.GetCodeHash(addr), polo.Bytes32{})

	assert.Nil(t, state.Err(), "error is not expected")

}

func TestStateRevert(t *testing.T) {
	kv, _ := storage.NewMem()
	state, _ := New(polo.Bytes32{}, kv)

	addr := polo.BytesToAddress([]byte("account1"))
	storageKey := polo.BytesToBytes32([]byte("storageKey"))

	values := []struct {
		balance *big.Int
		code    []byte
		storage polo.Bytes32
	}{
		{big.NewInt(1), []byte("code1"), polo.BytesToBytes32([]byte("v1"))},
		{big.NewInt(2), []byte("code2"), polo.BytesToBytes32([]byte("v2"))},
		{big.NewInt(3), []byte("code3"), polo.BytesToBytes32([]byte("v3"))},
	}

	var chk int
	for _, v := range values {
		chk = state.NewCheckpoint()
		state.SetBalance(addr, v.balance)
		state.SetCode(addr, v.code)
		state.SetStorage(addr, storageKey, v.storage)
	}

	for i := range values {
		v := values[len(values)-i-1]
		assert.Equal(t, state.GetBalance(addr), v.balance)
		assert.Equal(t, state.GetCode(addr), v.code)
		assert.Equal(t, state.GetCodeHash(addr), polo.Bytes32(crypto.Keccak256Hash(v.code)))
		assert.Equal(t, state.GetStorage(addr, storageKey), v.storage)
		state.RevertTo(chk)
		chk--
	}
	assert.False(t, state.Exists(addr))
	assert.Nil(t, state.Err(), "error is not expected")

	//
	state, _ = New(polo.Bytes32{}, kv)
	assert.Equal(t, state.NewCheckpoint(), 1)
	state.RevertTo(0)
	assert.Equal(t, state.NewCheckpoint(), 0)

}

func TestStorage(t *testing.T) {
	kv, _ := storage.NewMem()
	st, _ := New(polo.Bytes32{}, kv)

	addr := polo.BytesToAddress([]byte("addr"))
	key := polo.BytesToBytes32([]byte("key"))

	st.SetStorage(addr, key, polo.BytesToBytes32([]byte{1}))
	data, _ := rlp.EncodeToBytes([]byte{1})
	assert.Equal(t, rlp.RawValue(data), st.GetRawStorage(addr, key))

	st.SetRawStorage(addr, key, data)
	assert.Equal(t, polo.BytesToBytes32([]byte{1}), st.GetStorage(addr, key))

	st.SetStorage(addr, key, polo.Bytes32{})
	assert.Zero(t, len(st.GetRawStorage(addr, key)))

	v := struct {
		V1 uint
	}{313123}

	data, _ = rlp.EncodeToBytes(&v)
	st.SetRawStorage(addr, key, data)

	assert.Equal(t, polo.Blake2b(data), st.GetStorage(addr, key))
}