package subscriptions

import (
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/api/transactions"
)

type transactionReader struct {
	chain       *chain.Chain
	filter      *TranscationFilter
	blockReader chain.BlockReader
}

func newTransactionReader(chain *chain.Chain, position polo.Bytes32, filter *TranscationFilter) *transactionReader {
	return &transactionReader{
		chain:       chain,
		filter:      filter,
		blockReader: chain.NewBlockReader(position),
	}
}

func (tr *transactionReader) Read() ([]interface{}, bool, error) {
	blocks, err := tr.blockReader.Read()
	if err != nil {
		return nil, false, err
	}
	var msgs []interface{}
	for _, block := range blocks {
		txs := block.Transactions()
		for _, tx := range txs {
			if tr.filter.Match(tx) {
				txMeta, err := tr.chain.GetTransactionMeta(tx.ID(), block.Header().ID())
				if err != nil {
					return nil, false, err
				}

				receipt, err := tr.chain.GetTransactionReceipt(txMeta.BlockID, txMeta.Index)
				if err != nil {
					return nil, false, err
				}

				msg, err := transactions.ConvertReceipt(receipt, block.Header(), tx)
				if err != nil {
					return nil, false, err
				}
				msgs = append(msgs, msg)
			}
		}
	}
	return msgs, len(blocks) > 0, nil
}
