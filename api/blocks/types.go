// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package blocks

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/block"
)

//Block block
type Block struct {
	Number       uint32         `json:"number"`
	ID           polo.Bytes32   `json:"id"`
	Size         uint32         `json:"size"`
	ParentID     polo.Bytes32   `json:"parentID"`
	Timestamp    uint64         `json:"timestamp"`
	GasLimit     uint64         `json:"gasLimit"`
	Beneficiary  polo.Address   `json:"beneficiary"`
	GasUsed      uint64         `json:"gasUsed"`
	TotalScore   uint64         `json:"totalScore"`
	TxsRoot      polo.Bytes32   `json:"txsRoot"`
	StateRoot    polo.Bytes32   `json:"stateRoot"`
	ReceiptsRoot polo.Bytes32   `json:"receiptsRoot"`
	Signer       polo.Address   `json:"signer"`
	IsTrunk      bool           `json:"isTrunk"`
	Transactions []polo.Bytes32 `json:"transactions"`
}

func convertBlock(b *block.Block, isTrunk bool) (*Block, error) {
	if b == nil {
		return nil, nil
	}
	signer, err := b.Header().Signer()
	if err != nil {
		return nil, err
	}
	txs := b.Transactions()
	txIds := make([]polo.Bytes32, len(txs))
	for i, tx := range txs {
		txIds[i] = tx.ID()
	}

	header := b.Header()
	return &Block{
		Number:       header.Number(),
		ID:           header.ID(),
		ParentID:     header.ParentID(),
		Timestamp:    header.Timestamp(),
		TotalScore:   header.TotalScore(),
		GasLimit:     header.GasLimit(),
		GasUsed:      header.GasUsed(),
		Beneficiary:  header.Beneficiary(),
		Signer:       signer,
		Size:         uint32(b.Size()),
		StateRoot:    header.StateRoot(),
		ReceiptsRoot: header.ReceiptsRoot(),
		TxsRoot:      header.TxsRoot(),
		IsTrunk:      isTrunk,
		Transactions: txIds,
	}, nil
}
