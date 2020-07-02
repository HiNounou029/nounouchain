// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package block

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/tx"
)

// Builder to make it easy to build a block object.
type Builder struct {
	headerBody headerBody
	txs        tx.Transactions
}

// ParentID set parent id.
func (b *Builder) ParentID(id polo.Bytes32) *Builder {
	b.headerBody.ParentID = id
	return b
}

// Timestamp set timestamp.
func (b *Builder) Timestamp(ts uint64) *Builder {
	b.headerBody.Timestamp = ts
	return b
}

// TotalScore set total score.
func (b *Builder) TotalScore(score uint64) *Builder {
	b.headerBody.TotalScore = score
	return b
}

// GasLimit set gas limit.
func (b *Builder) GasLimit(limit uint64) *Builder {
	b.headerBody.GasLimit = limit
	return b
}

// GasUsed set gas used.
func (b *Builder) GasUsed(used uint64) *Builder {
	b.headerBody.GasUsed = used
	return b
}

// Beneficiary set recipient of reward.
func (b *Builder) Beneficiary(addr polo.Address) *Builder {
	b.headerBody.Beneficiary = addr
	return b
}

// StateRoot set state root.
func (b *Builder) StateRoot(hash polo.Bytes32) *Builder {
	b.headerBody.StateRoot = hash
	return b
}

// ReceiptsRoot set receipts root.
func (b *Builder) ReceiptsRoot(hash polo.Bytes32) *Builder {
	b.headerBody.ReceiptsRoot = hash
	return b
}

// Transaction add a transaction.
func (b *Builder) Transaction(tx *tx.Transaction) *Builder {
	b.txs = append(b.txs, tx)
	return b
}

// Build build a block object.
func (b *Builder) Build() *Block {
	header := Header{body: b.headerBody}
	header.body.TxsRoot = b.txs.RootHash()

	return &Block{
		header: &header,
		txs:    b.txs,
	}
}
