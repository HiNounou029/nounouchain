// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package txpool

import (
	"testing"
	"time"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/genesis"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/tx"
	Tx "github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/inconshreveable/log15"
	"github.com/stretchr/testify/assert"
)

func init() {
	log15.Root().SetHandler(log15.DiscardHandler())
}

func newPool() *TxPool {
	kv, _ := storage.NewMem()
	chain := newChain(kv)
	return New(chain, state.NewCreator(kv), Options{
		Limit:           10,
		LimitPerAccount: 2,
		MaxLifetime:     time.Hour,
	})
}
func TestNewClose(t *testing.T) {
	pool := newPool()
	defer pool.Close()
}

func TestSubscribeNewTx(t *testing.T) {
	pool := newPool()
	defer pool.Close()

	b1 := new(block.Builder).
		ParentID(pool.chain.GenesisBlock().Header().ID()).
		Timestamp(uint64(time.Now().Unix())).
		TotalScore(100).
		GasLimit(10000000).
		StateRoot(pool.chain.GenesisBlock().Header().StateRoot()).
		Build()
	pool.chain.AddBlock(b1, nil)

	txCh := make(chan *TxEvent)

	pool.SubscribeTxEvent(txCh)

	tx := newTx(pool.chain.Tag(), nil, 21000, tx.BlockRef{}, 100, nil, genesis.DevAccounts()[0])
	assert.Nil(t, pool.Add(tx))

	v := true
	assert.Equal(t, &TxEvent{tx, &v}, <-txCh)
}

func TestWashTxs(t *testing.T) {
	pool := newPool()
	defer pool.Close()
	txs, _, err := pool.wash(pool.chain.BestBlock().Header())
	assert.Nil(t, err)
	assert.Zero(t, len(txs))
	assert.Zero(t, len(pool.Executables()))

	tx := newTx(pool.chain.Tag(), nil, 21000, tx.BlockRef{}, 100, nil, genesis.DevAccounts()[0])
	assert.Nil(t, pool.Add(tx))

	txs, _, err = pool.wash(pool.chain.BestBlock().Header())
	assert.Nil(t, err)
	assert.Equal(t, Tx.Transactions{tx}, txs)

	b1 := new(block.Builder).
		ParentID(pool.chain.GenesisBlock().Header().ID()).
		Timestamp(uint64(time.Now().Unix())).
		TotalScore(100).
		GasLimit(10000000).
		StateRoot(pool.chain.GenesisBlock().Header().StateRoot()).
		Build()
	pool.chain.AddBlock(b1, nil)

	txs, _, err = pool.wash(pool.chain.BestBlock().Header())
	assert.Nil(t, err)
	assert.Equal(t, Tx.Transactions{tx}, txs)
}

func TestAdd(t *testing.T) {
	pool := newPool()
	defer pool.Close()
	b1 := new(block.Builder).
		ParentID(pool.chain.GenesisBlock().Header().ID()).
		Timestamp(uint64(time.Now().Unix())).
		TotalScore(100).
		GasLimit(10000000).
		StateRoot(pool.chain.GenesisBlock().Header().StateRoot()).
		Build()
	pool.chain.AddBlock(b1, nil)
	acc := genesis.DevAccounts()[0]

	dupTx := newTx(pool.chain.Tag(), nil, 21000, tx.BlockRef{}, 100, nil, acc)

	tests := []struct {
		tx     *tx.Transaction
		errStr string
	}{
		{newTx(pool.chain.Tag()+1, nil, 21000, tx.BlockRef{}, 100, nil, acc), "bad tx: chain tag mismatch"},
		{dupTx, ""},
		{dupTx, ""},
	}

	for _, tt := range tests {
		err := pool.Add(tt.tx)
		if tt.errStr == "" {
			assert.Nil(t, err)
		} else {
			assert.Contains(t, err.Error(), tt.errStr)
			//assert.Equal(t, tt.errStr, err.Error())
		}
	}

	tests = []struct {
		tx     *tx.Transaction
		errStr string
	}{
		{newTx(pool.chain.Tag(), nil, 21000, tx.NewBlockRef(200), 100, nil, acc), "tx rejected: tx is not executable"},
		{newTx(pool.chain.Tag(), nil, 21000, tx.BlockRef{}, 100, &polo.Bytes32{1}, acc), "tx rejected: tx is not executable"},
	}

	for _, tt := range tests {
		err := pool.StrictlyAdd(tt.tx)
		if tt.errStr == "" {
			assert.Nil(t, err)
		} else {
			assert.Equal(t, tt.errStr, err.Error())
		}
	}
}
