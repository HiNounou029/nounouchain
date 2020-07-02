// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package params

import (
	"math/big"
	"testing"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/stretchr/testify/assert"
)

func TestParamsGetSet(t *testing.T) {
	kv, _ := storage.NewMem()
	st, _ := state.New(polo.Bytes32{}, kv)
	setv := big.NewInt(10)
	key := polo.BytesToBytes32([]byte("key"))
	p := New(polo.BytesToAddress([]byte("par")), st)
	p.Set(key, setv)

	getv := p.Get(key)
	assert.Equal(t, setv, getv)

	assert.Nil(t, st.Err())
}
