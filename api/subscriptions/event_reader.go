// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package subscriptions

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/chain"
)

type eventReader struct {
	chain       *chain.Chain
	filter      *EventFilter
	blockReader chain.BlockReader
}

func newEventReader(chain *chain.Chain, position polo.Bytes32, filter *EventFilter) *eventReader {
	return &eventReader{
		chain:       chain,
		filter:      filter,
		blockReader: chain.NewBlockReader(position),
	}
}

func (er *eventReader) Read() ([]interface{}, bool, error) {
	blocks, err := er.blockReader.Read()
	if err != nil {
		return nil, false, err
	}
	var msgs []interface{}
	for _, block := range blocks {
		receipts, err := er.chain.GetBlockReceipts(block.Header().ID())
		if err != nil {
			return nil, false, err
		}
		txs := block.Transactions()
		for i, receipt := range receipts {
			for _, output := range receipt.Outputs {
				for _, event := range output.Events {
					if er.filter.Match(event) {
						msg, err := convertEvent(block.Header(), txs[i], event, block.Obsolete)
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
