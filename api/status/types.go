// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package status

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/chain"
)

type ChainStatus struct {
	Tag             byte         `json:"tag"`
	BestBlockNumber uint32       `json:"bestBlockNum"`
	BestBlockId     polo.Bytes32 `json:"bestBlockId"`
}

func convertChainStatus(chain *chain.Chain) *ChainStatus {
	status := &ChainStatus{
		Tag:             chain.Tag(),
		BestBlockId:     chain.BestBlock().Header().ID(),
		BestBlockNumber: chain.BestBlock().Header().Number(),
	}
	return status
}
