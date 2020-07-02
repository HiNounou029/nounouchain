// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package main

import (
	"fmt"
	"github.com/HiNounou029/nounouchain/cmd/bridge"
	"github.com/HiNounou029/nounouchain/cmd/bridge/contract"
	"github.com/pkg/errors"
	"math/big"
	"os"
	"time"
)

var chainMap *bridge.ChainMap

func forwardRequest(srcChain *bridge.ChainInfo, msgId int64,
	destChainId *big.Int, dstBrg *contract.BridgeContract) error {
	srcBrg := srcChain.Bridge
	req, err := srcBrg.SendMsgsByDestChain(destChainId, big.NewInt(msgId-1))
	if err != nil {
		return errors.Wrapf(err, "failed to get message %s/%d at chain %s",
			destChainId, msgId, srcChain.NetworkID)
	}
	txId, err := dstBrg.Forward(srcChain.NetworkID, big.NewInt(msgId), req)
	if err != nil {
		return err
	}
	fmt.Printf("forwarded request %s->%s %d %s\n",
		srcChain.NetworkID, destChainId, msgId, txId)
	return nil
}

// forward response of <srcChain, msgId, destChain> from dsgBrg to srcChain
func forwardResponse(srcChain *bridge.ChainInfo, msgId int64,
	destChainId *big.Int, dstBrg *contract.BridgeContract) error {
	res, err := dstBrg.RecvMsgsBySrcChain(srcChain.NetworkID, big.NewInt(msgId-1))
	if err != nil {
		return errors.Wrapf(err, "failed to get response message for %s/%d at %s",
			srcChain.NetworkID, msgId, destChainId)
	}
	srcBrg := srcChain.Bridge
	txId, err := srcBrg.Ack(destChainId, big.NewInt(msgId), res.Success, res.ResMsg)
	if err != nil {
		return errors.Wrapf(err, "failed to ack message %s/%d",
			srcChain.NetworkID, msgId)
	}
	fmt.Printf("forwarded response %s->%s %d %s\n",
		srcChain.NetworkID, destChainId, msgId, txId)
	return nil
}

// syncMsg from srcChain/msgId to destChainId, forward as needed,
// return true if ack is forwarded
func syncMsg(srcChain *bridge.ChainInfo, destChainId *big.Int, msgId int64) (bool, error) {
	fmt.Printf("synching message %s -> %s msgId: %d\n",
		srcChain.NetworkID.String(), destChainId.String(), msgId)
	// 检查destChain的lastMsgIdBySrcChain，已确认msgId是否已经发送过了，如果没有发送过，则发送
	dstBrg := chainMap.ById(destChainId).Bridge
	lastRecvMsgId, err := dstBrg.LastMsgIdBySrcChain(srcChain.NetworkID)
	if err != nil {
		return false, err
	}
	fmt.Printf("lastRecvMsgId %s in chain %s by srcChain %s\n",
		lastRecvMsgId, destChainId, srcChain.NetworkID)
	if lastRecvMsgId.Int64() < msgId {
		// 还没有发送过
		err := forwardRequest(srcChain, msgId, destChainId, dstBrg)
		return false, err
	} else {
		// 发送过了，转发响应
		err := forwardResponse(srcChain, msgId, destChainId, dstBrg)
		return true, err
	}
}

// sync messages from srcChain to destChainId
// return true if all messages are processed and acked back to srcChain
func syncChainPair(srcChain *bridge.ChainInfo, destChainId *big.Int) (bool, error) {
	fmt.Printf("synching chain %s -> %s\n", srcChain.NetworkID, destChainId)
	brg := srcChain.Bridge

	lastAckId, _ := brg.LastAckedMsgIdByDestChain(destChainId)
	lastSentId, _ := brg.LastSentMsgIdByDestChain(destChainId)

	if lastAckId.Cmp(lastSentId) >= 0 {
		// all messages are acked
		return true, nil
	}

	var allAcked = true

	for id := lastAckId.Int64() + 1; id <= lastSentId.Int64(); id++ {
		ack, err := syncMsg(srcChain, destChainId, id)
		if err != nil {
			return false, err
		}
		if !ack {
			allAcked = false
		}
	}

	return allAcked, nil
}

func syncChain(chain *bridge.ChainInfo, allChainIds []*big.Int) error {
	for _, chainId := range allChainIds {
		if chain.NetworkID.Cmp(chainId) != 0 {
			var err error = nil
			synced, err := syncChainPair(chain, chainId)
			if err != nil {
				fmt.Printf("sync chain pair error %s->%s %+v\n",
					chain.NetworkID, chainId.String(), err)
				return err
			}
			if synced {
				fmt.Printf("chain %s->%s is synced\n", chain.NetworkID, chainId.String())
			}
		}
	}
	return nil
}

func syncChains() {
	chainIds := chainMap.ChainIds()

	for _, chain := range chainMap.Chains() {
		err := syncChain(chain, chainIds)
		if err != nil {
			fmt.Printf("syncChain %s error\n", chain.NetworkID)
			return
		}
	}
}

func syncChainLoop() {
	for true {
		syncChains()
		time.Sleep(time.Second * 5)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("missing config file path")
		return
	}
	configFilePath := os.Args[1]
	var err error
	chainMap, err = bridge.InitChains(configFilePath, false)

	if err != nil {
		fmt.Printf("init chains: %v", err)
		return
	}

	syncChainLoop()
}
