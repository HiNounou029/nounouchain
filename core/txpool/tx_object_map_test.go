// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package txpool

import (
	"errors"
	"testing"

	"github.com/HiNounou029/nounouchain/nounou/genesis"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/stretchr/testify/assert"
)

func TestTxObjMap(t *testing.T) {

	kv, _ := storage.NewMem()
	chain := newChain(kv)

	tx1 := newTx(chain.Tag(), nil, 21000, tx.BlockRef{}, 100, nil, genesis.DevAccounts()[0])
	tx2 := newTx(chain.Tag(), nil, 21000, tx.BlockRef{}, 100, nil, genesis.DevAccounts()[0])
	tx3 := newTx(chain.Tag(), nil, 21000, tx.BlockRef{}, 100, nil, genesis.DevAccounts()[1])

	txObj1, _ := resolveTx(tx1)
	txObj2, _ := resolveTx(tx2)
	txObj3, _ := resolveTx(tx3)

	m := newTxObjectMap()
	assert.Zero(t, m.Len())

	assert.Nil(t, m.Add(txObj1, 1))
	assert.Nil(t, m.Add(txObj1, 1), "should no error if exists")
	assert.Equal(t, 1, m.Len())

	assert.Equal(t, errors.New("account quota exceeded"), m.Add(txObj2, 1))
	assert.Equal(t, 1, m.Len())

	assert.Nil(t, m.Add(txObj3, 1))
	assert.Equal(t, 2, m.Len())

	assert.True(t, m.Contains(tx1.ID()))
	assert.False(t, m.Contains(tx2.ID()))
	assert.True(t, m.Contains(tx3.ID()))

	assert.True(t, m.Remove(tx1.ID()))
	assert.False(t, m.Contains(tx1.ID()))
	assert.False(t, m.Remove(tx2.ID()))

	assert.Equal(t, []*txObject{txObj3}, m.ToTxObjects())
	assert.Equal(t, tx.Transactions{tx3}, m.ToTxs())

}
