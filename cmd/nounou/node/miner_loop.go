// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package node

import (
	"context"
	"fmt"
	"time"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/miner"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/mclock"
	"github.com/pkg/errors"
)

func (n *Node) minerLoop(ctx context.Context) {
	log.Debug("enter miner loop")
	defer log.Debug("leave miner loop")

	log.Info("waiting for synchronization...")
	select {
	case <-ctx.Done():
		return
	case <-n.comm.Synced():
	}
	log.Info("synchronization process done")

	var (
		authorized bool
		flow       *miner.Flow
		err        error
		ticker     = time.NewTicker(time.Second)
	)
	defer ticker.Stop()

	nopackcount := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		best := n.chain.BestBlock()
		now := uint64(time.Now().Unix())

		if flow == nil {
			if flow, err = n.miner.Schedule(best.Header(), now); err != nil {
				if authorized {
					authorized = false
					log.Warn("unable to mine block", "err", err)
				}
				nopackcount = 0
				continue
			}
			if !authorized {
				authorized = true
				log.Info("prepared to mine block")
			}
			log.Debug("scheduled to mine block", "after", time.Duration(flow.When()-now)*time.Second)
			continue
		}

		if flow.ParentHeader().ID() != best.Header().ID() {
			flow = nil
			nopackcount++
			log.Debug("re-schedule miner due to new best block")
			continue
		}

		if nopackcount >= 10 {
			time.Sleep(time.Duration(polo.Conf.BlockInterval-1)* time.Second)
		}

		now = uint64(time.Now().Unix())
//		fmt.Println("now:", now, "flow:", flow.When(), "nopackcount:", nopackcount)

		if now+1 >= flow.When() {
			if err := n.pack(flow); err != nil {
				log.Error("failed to create block", "err", err)
			}
			flow = nil
			nopackcount = 0
		}
	}
}

func (n *Node) pack(flow *miner.Flow) error {
	txs := n.txPool.Executables()
	var txsToRemove []polo.Bytes32
	defer func() {
		for _, id := range txsToRemove {
			n.txPool.Remove(id)
		}
	}()

	log.Info("txpool monitor: ", "len", len(txs))

	startTime := mclock.Now()
	var count uint64
	var MAXTxs = polo.Conf.BlockInterval * polo.Conf.TxPerSecondLimit
	for _, tx := range txs {
		if err := flow.Adopt(tx); err != nil {
			if miner.IsGasLimitReached(err) {
				break
			}
			if miner.IsTxNotAdoptableNow(err) {
				continue
			}
			txsToRemove = append(txsToRemove, tx.ID())
		} else {
			count++
			if count >= MAXTxs {
				break
			}
		}
	}

	newBlock, stage, receipts, err := flow.Pack(n.master.PrivateKey)
	if err != nil {
		return err
	}
	execElapsed := mclock.Now() - startTime

	if _, err := stage.Commit(); err != nil {
		return errors.WithMessage(err, "commit state")
	}

	fork, err := n.commitBlock(newBlock, receipts)
	if err != nil {
		return errors.WithMessage(err, "commit block")
	}
	commitElapsed := mclock.Now() - startTime - execElapsed

	n.processFork(fork)

	if len(fork.Trunk) > 0 {
		n.comm.BroadcastBlock(newBlock)
		log.Info("new block mined: ",
			"txs", len(receipts),
			"mgas", float64(newBlock.Header().GasUsed())/1000/1000,
			"et", fmt.Sprintf("%v|%v", common.PrettyDuration(execElapsed), common.PrettyDuration(commitElapsed)),
			"id", shortID(newBlock.Header().ID()),
		)
	}

	n.miner.SetTargetGasLimit(0)
	if execElapsed > 0 {
		gasUsed := newBlock.Header().GasUsed()
		// calc target gas limit only if gas used above third of gas limit
		if gasUsed > newBlock.Header().GasLimit()/3 {
			targetGasLimit := uint64(polo.TolerableBlockPackingTime) * gasUsed / uint64(execElapsed)
			n.miner.SetTargetGasLimit(targetGasLimit)
			log.Debug("reset target gas limit", "value", targetGasLimit)
		}
	}
	return nil
}
