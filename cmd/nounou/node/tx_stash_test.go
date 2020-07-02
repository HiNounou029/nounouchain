// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package node

import (
	"bytes"
	"math/rand"
	"sort"
	"testing"

	"github.com/HiNounou029/nounouchain/nounou/genesis"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/stretchr/testify/assert"
)

func newTx() *tx.Transaction {
	tx := new(tx.Builder).Nonce(rand.Uint64()).Build()
	sig, _ := crypto.Sign(tx.SigningHash().Bytes(), genesis.DevAccounts()[0].PrivateKey)
	return tx.WithSignature(sig)
}

func TestTxStash(t *testing.T) {
	db, _ := storage.NewMem()
	defer db.Close()

	stash := newTxStash(db, 10)

	var saved tx.Transactions
	for i := 0; i < 11; i++ {
		tx := newTx()
		assert.Nil(t, stash.Save(tx))
		saved = append(saved, tx)
	}

	loaded := newTxStash(db, 10).LoadAll()

	saved = saved[1:]
	sort.Slice(saved, func(i, j int) bool {
		return bytes.Compare(saved[i].ID().Bytes(), saved[j].ID().Bytes()) < 0
	})

	sort.Slice(loaded, func(i, j int) bool {
		return bytes.Compare(loaded[i].ID().Bytes(), loaded[j].ID().Bytes()) < 0
	})

	assert.Equal(t, saved.RootHash(), loaded.RootHash())
}
