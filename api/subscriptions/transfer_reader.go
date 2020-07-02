// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package subscriptions

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/chain"
)

type transferReader struct {
	chain       *chain.Chain
	filter      *TransferFilter
	blockReader chain.BlockReader
}

func newTransferReader(chain *chain.Chain, position polo.Bytes32, filter *TransferFilter) *transferReader {
	return &transferReader{
		chain:       chain,
		filter:      filter,
		blockReader: chain.NewBlockReader(position),
	}
}

func (tr *transferReader) Read() ([]interface{}, bool, error) {
	blocks, err := tr.blockReader.Read()
	if err != nil {
		return nil, false, err
	}
	var msgs []interface{}
	for _, block := range blocks {
		receipts, err := tr.chain.GetBlockReceipts(block.Header().ID())
		if err != nil {
			return nil, false, err
		}
		txs := block.Transactions()
		for i, receipt := range receipts {
			for _, output := range receipt.Outputs {
				for _, transfer := range output.Transfers {
					origin, err := txs[i].Signer()
					if err != nil {
						return nil, false, err
					}
					if tr.filter.Match(transfer, origin) {
						msg, err := convertTransfer(block.Header(), txs[i], transfer, block.Obsolete)
						if err != nil {
							return nil, false, err
						}
						msgs = append(msgs, msg)
					}
				}
			}
		}
	}
	return msgs, len(blocks) > 0, nil
}
