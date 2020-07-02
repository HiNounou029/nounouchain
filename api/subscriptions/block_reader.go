// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package subscriptions

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/chain"
)

type blockReader struct {
	chain       *chain.Chain
	blockReader chain.BlockReader
}

func newBlockReader(chain *chain.Chain, position polo.Bytes32) *blockReader {
	return &blockReader{
		chain:       chain,
		blockReader: chain.NewBlockReader(position),
	}
}

func (br *blockReader) Read() ([]interface{}, bool, error) {
	blocks, err := br.blockReader.Read()
	if err != nil {
		return nil, false, err
	}
	var msgs []interface{}
	for _, block := range blocks {
		msg, err := convertBlock(block)
		if err != nil {
			return nil, false, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, len(blocks) > 0, nil
}
