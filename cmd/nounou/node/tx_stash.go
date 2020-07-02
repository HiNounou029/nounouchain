// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package node

import (
	"container/list"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/storage/kv"
	"github.com/ethereum/go-ethereum/rlp"
)

// to stash non-executable txs.
// it uses a FIFO queue to limit the size of stash.
type txStash struct {
	kv      kv.GetPutter
	fifo    *list.List
	maxSize int
}

func newTxStash(kv kv.GetPutter, maxSize int) *txStash {
	return &txStash{kv, list.New(), maxSize}
}

func (ts *txStash) Save(tx *tx.Transaction) error {
	has, err := ts.kv.Has(tx.ID().Bytes())
	if err != nil {
		return err
	}
	if has {
		return nil
	}

	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return err
	}

	if err := ts.kv.Put(tx.ID().Bytes(), data); err != nil {
		return err
	}
	ts.fifo.PushBack(tx.ID())
	for ts.fifo.Len() > ts.maxSize {
		keyToDelete := ts.fifo.Remove(ts.fifo.Front()).(polo.Bytes32).Bytes()
		if err := ts.kv.Delete(keyToDelete); err != nil {
			return err
		}
	}
	return nil
}

func (ts *txStash) LoadAll() tx.Transactions {
	var txs tx.Transactions
	iter := ts.kv.NewIterator(*kv.NewRangeWithBytesPrefix(nil))
	for iter.Next() {
		var tx tx.Transaction
		if err := rlp.DecodeBytes(iter.Value(), &tx); err != nil {
			log.Warn("decode stashed tx", "err", err)
			if err := ts.kv.Delete(iter.Key()); err != nil {
				log.Warn("delete corrupted stashed tx", "err", err)
			}
		} else {
			txs = append(txs, &tx)
			ts.fifo.PushBack(tx.ID())
		}
	}
	return txs
}
