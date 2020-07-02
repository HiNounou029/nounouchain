// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package comm

import (
	"github.com/HiNounou029/nounouchain/core/txpool"
	"github.com/HiNounou029/nounouchain/network/comm/proto"
)

func (c *Communicator) txsLoop() {

	txEvCh := make(chan *txpool.TxEvent, 10)
	sub := c.txPool.SubscribeTxEvent(txEvCh)
	defer sub.Unsubscribe()

	for {
		select {
		case <-c.ctx.Done():
			return
		case txEv := <-txEvCh:
			if txEv.Executable != nil && *txEv.Executable {
				tx := txEv.Tx
				peers := c.peerSet.Slice().Filter(func(p *Peer) bool {
					return !p.IsTransactionKnown(tx.ID())
				})

				for _, peer := range peers {
					peer := peer
					peer.MarkTransaction(tx.ID())
					c.goes.Go(func() {
						if err := proto.NotifyNewTx(c.ctx, peer, tx); err != nil {
							peer.logger.Debug("failed to broadcast tx", "err", err)
						}
					})
				}
			}
		}
	}
}
