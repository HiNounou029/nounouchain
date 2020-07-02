// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package block

import (
	"bytes"

	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/ethereum/go-ethereum/rlp"
)

// Raw allows to partially decode components of a block.
type Raw []byte

// DecodeHeader decode only the header.
func (r Raw) DecodeHeader() (*Header, error) {
	content, _, err := rlp.SplitList(r)
	if err != nil {
		return nil, err
	}

	var header Header
	if err := rlp.Decode(bytes.NewReader(content), &header); err != nil {
		return nil, err
	}
	return &header, nil
}

// DecodeBody decode only the body.
func (r Raw) DecodeBody() (*Body, error) {
	content, _, err := rlp.SplitList(r)
	if err != nil {
		return nil, err
	}

	_, _, rest, err := rlp.Split(content)
	if err != nil {
		return nil, err
	}
	var txs tx.Transactions
	if err := rlp.Decode(bytes.NewReader(rest), &txs); err != nil {
		return nil, err
	}
	return &Body{txs}, nil
}
