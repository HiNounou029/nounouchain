// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package tx

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/trie"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	emptyRoot = trie.DeriveRoot(&derivableTxs{})
)

// Transactions a slice of transactions.
type Transactions []*Transaction

// RootHash computes merkle root hash of transactions.
func (txs Transactions) RootHash() polo.Bytes32 {
	if len(txs) == 0 {
		// optimized
		return emptyRoot
	}
	return trie.DeriveRoot(derivableTxs(txs))
}

// implements types.DerivableList
type derivableTxs Transactions

func (txs derivableTxs) Len() int {
	return len(txs)
}

func (txs derivableTxs) GetRlp(i int) []byte {
	data, err := rlp.EncodeToBytes(txs[i])
	if err != nil {
		panic(err)
	}
	return data
}
