// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package miner

import (
	"crypto/ecdsa"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/vm/runtime"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/pkg/errors"
)

// Flow the flow of packing a new block.
type Flow struct {
	miner        *Miner
	parentHeader *block.Header
	runtime      *runtime.Runtime
	processedTxs map[polo.Bytes32]bool // txID -> reverted
	gasUsed      uint64
	txs          tx.Transactions
	receipts     tx.Receipts
}

func newFlow(
	miner *Miner,
	parentHeader *block.Header,
	runtime *runtime.Runtime,
) *Flow {
	return &Flow{
		miner:        miner,
		parentHeader: parentHeader,
		runtime:      runtime,
		processedTxs: make(map[polo.Bytes32]bool),
	}
}

// ParentHeader returns parent block header.
func (f *Flow) ParentHeader() *block.Header {
	return f.parentHeader
}

// When the target time to do packing.
func (f *Flow) When() uint64 {
	return f.runtime.Context().Time
}

func (f *Flow) findTx(txID polo.Bytes32) (found bool, reverted bool, err error) {
	if reverted, ok := f.processedTxs[txID]; ok {
		return true, reverted, nil
	}
	txMeta, err := f.miner.chain.GetTransactionMeta(txID, f.parentHeader.ID())
	if err != nil {
		if f.miner.chain.IsNotFound(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, txMeta.Reverted, nil
}

// Adopt try to execute the given transaction.
// If the tx is valid and can be executed on current state (regardless of VM error),
// it will be adopted by the new block.
func (f *Flow) Adopt(tx *tx.Transaction) error {
	switch {
	case tx.ChainTag() != f.miner.chain.Tag():
		return badTxError{"chain tag mismatch"}
	case tx.HasReservedFields():
		return badTxError{"reserved fields not empty"}
	case f.runtime.Context().Number < tx.BlockRef().Number():
		return errTxNotAdoptableNow
	case tx.IsExpired(f.runtime.Context().Number):
		return badTxError{"expired"}
	case f.gasUsed+tx.Gas() > f.runtime.Context().GasLimit:
		// gasUsed < 90% gas limit
		if float64(f.gasUsed)/float64(f.runtime.Context().GasLimit) < 0.9 {
			// try to find a lower gas tx
			return errTxNotAdoptableNow
		}
		return errGasLimitReached
	}

	// check if tx already there
	if found, _, err := f.findTx(tx.ID()); err != nil {
		return err
	} else if found {
		return errKnownTx
	}

	if dependsOn := tx.DependsOn(); dependsOn != nil {
		// check if deps exists
		found, reverted, err := f.findTx(*dependsOn)
		if err != nil {
			return err
		}
		if !found {
			return errTxNotAdoptableNow
		}
		if reverted {
			return errTxNotAdoptableForever
		}
	}

	checkpoint := f.runtime.State().NewCheckpoint()
	receipt, err := f.runtime.ExecuteTransaction(tx)
	if err != nil {
		// skip and revert state
		f.runtime.State().RevertTo(checkpoint)
		return badTxError{err.Error()}
	}
	f.processedTxs[tx.ID()] = receipt.Reverted
	f.gasUsed += receipt.GasUsed
	f.receipts = append(f.receipts, receipt)
	f.txs = append(f.txs, tx)
	return nil
}

// Pack build and sign the new block.
func (f *Flow) Pack(privateKey *ecdsa.PrivateKey) (*block.Block, *state.Stage, tx.Receipts, error) {
	if f.miner.nodeMaster != polo.Address(crypto.PubkeyToAddress(privateKey.PublicKey)) {
		return nil, nil, nil, errors.New("private key mismatch")
	}

	if err := f.runtime.Seeker().Err(); err != nil {
		return nil, nil, nil, err
	}

	stage := f.runtime.State().Stage()
	stateRoot, err := stage.Hash()
	if err != nil {
		return nil, nil, nil, err
	}

	builder := new(block.Builder).
		Beneficiary(f.runtime.Context().Beneficiary).
		GasLimit(f.runtime.Context().GasLimit).
		ParentID(f.parentHeader.ID()).
		Timestamp(f.runtime.Context().Time).
		TotalScore(f.runtime.Context().TotalScore).
		GasUsed(f.gasUsed).
		ReceiptsRoot(f.receipts.RootHash()).
		StateRoot(stateRoot)
	for _, tx := range f.txs {
		builder.Transaction(tx)
	}
	newBlock := builder.Build()

	sig, err := crypto.Sign(newBlock.Header().SigningHash().Bytes(), privateKey)
	if err != nil {
		return nil, nil, nil, err
	}
	return newBlock.WithSignature(sig), stage, f.receipts, nil
}
