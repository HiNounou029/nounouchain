// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package txpool

import (
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/genesis"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/HiNounou029/nounouchain/storage/kv"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/stretchr/testify/assert"
)

func newChain(kv kv.GetPutter) *chain.Chain {
	gene := genesis.NewDevnet()
	b0, _, _ := gene.Build(state.NewCreator(kv))
	chain, _ := chain.New(kv, b0)
	return chain
}

func signTx(tx *tx.Transaction, acc genesis.DevAccount) *tx.Transaction {
	sig, _ := crypto.Sign(tx.SigningHash().Bytes(), acc.PrivateKey)
	return tx.WithSignature(sig)
}

func newTx(chainTag byte, clauses []*tx.Clause, gas uint64, blockRef tx.BlockRef, expiration uint32, dependsOn *polo.Bytes32, from genesis.DevAccount) *tx.Transaction {
	builder := new(tx.Builder).ChainTag(chainTag)
	for _, c := range clauses {
		builder.Clause(c)
	}

	tx := builder.BlockRef(blockRef).
		Expiration(expiration).
		Nonce(rand.Uint64()).
		DependsOn(dependsOn).
		Gas(gas).Build()

	return signTx(tx, from)
}

func TestSort(t *testing.T) {
	objs := []*txObject{
		{overallGasPrice: big.NewInt(10)},
		{overallGasPrice: big.NewInt(20)},
		{overallGasPrice: big.NewInt(30)},
	}
	sortTxObjsByOverallGasPriceDesc(objs)

	assert.Equal(t, big.NewInt(30), objs[0].overallGasPrice)
	assert.Equal(t, big.NewInt(20), objs[1].overallGasPrice)
	assert.Equal(t, big.NewInt(10), objs[2].overallGasPrice)
}

func TestResolve(t *testing.T) {
	acc := genesis.DevAccounts()[0]
	tx := newTx(0, nil, 21000, tx.BlockRef{}, 100, nil, acc)

	txObj, err := resolveTx(tx)
	assert.Nil(t, err)
	assert.Equal(t, tx, txObj.Transaction)

	assert.Equal(t, acc.Address, txObj.Origin())

}

func TestExecutable(t *testing.T) {
	acc := genesis.DevAccounts()[0]

	kv, _ := storage.NewMem()
	chain := newChain(kv)
	b0 := chain.GenesisBlock()
	b1 := new(block.Builder).ParentID(b0.Header().ID()).GasLimit(10000000).TotalScore(100).Build()
	chain.AddBlock(b1, nil)
	st, _ := state.New(chain.GenesisBlock().Header().StateRoot(), kv)

	tests := []struct {
		tx          *tx.Transaction
		expected    bool
		expectedErr string
	}{
		{newTx(0, nil, 21000, tx.BlockRef{}, 100, nil, acc), true, ""},
		{newTx(0, nil, math.MaxUint64, tx.BlockRef{}, 100, nil, acc), false, "gas too large"},
		{newTx(0, nil, 21000, tx.BlockRef{1}, 100, nil, acc), true, "block ref out of schedule"},
		//{newTx(0, nil, 21000, tx.BlockRef{0}, 0, nil, acc), true, "expired"},
		{newTx(0, nil, 21000, tx.BlockRef{0}, 100, &polo.Bytes32{}, acc), false, ""},
	}

	for _, tt := range tests {
		txObj, err := resolveTx(tt.tx)
		assert.Nil(t, err)

		exe, err := txObj.Executable(chain, st, b1.Header())
		if tt.expectedErr != "" {
			if err == nil {
				t.Fatalf("want error %s, actual error is nil", tt.expectedErr)
			}
			assert.Equal(t, tt.expectedErr, err.Error())
		} else {
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, exe)
		}
	}
}
