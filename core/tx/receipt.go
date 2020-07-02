// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package tx

import (
	"math/big"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/trie"
	"github.com/ethereum/go-ethereum/rlp"
)

// Receipt represents the results of a transaction.
type Receipt struct {
	// gas used by this tx
	GasUsed uint64
	// the one who paid for gas
	GasPayer polo.Address
	// energy paid for used gas
	Paid *big.Int
	// gas reward given to block proposer
	Reward *big.Int
	// if the tx reverted
	Reverted bool
	// outputs of clauses in tx
	Outputs []*Output
}

// Output output of clause execution.
type Output struct {
	// events produced by the clause
	Events Events
	// transfer occurred in clause
	Transfers Transfers
}

// Receipts slice of receipts.
type Receipts []*Receipt

// RootHash computes merkle root hash of receipts.
func (rs Receipts) RootHash() polo.Bytes32 {
	if len(rs) == 0 {
		// optimized
		return emptyRoot
	}
	return trie.DeriveRoot(derivableReceipts(rs))
}

// implements DerivableList
type derivableReceipts Receipts

func (rs derivableReceipts) Len() int {
	return len(rs)
}
func (rs derivableReceipts) GetRlp(i int) []byte {
	data, err := rlp.EncodeToBytes(rs[i])
	if err != nil {
		panic(err)
	}
	return data
}
