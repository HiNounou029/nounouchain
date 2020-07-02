// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package node

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/network/comm"
)

type Network interface {
	PeersStats() []*comm.PeerStats
}

type PeerStats struct {
	Name        string       `json:"name"`
	BestBlockID polo.Bytes32 `json:"bestBlockID"`
	TotalScore  uint64       `json:"totalScore"`
	PeerID      string       `json:"peerID"`
	NetAddr     string       `json:"netAddr"`
	Inbound     bool         `json:"inbound"`
	Duration    uint64       `json:"duration"`
	PeerAddress string       `json:"peerAddr"`
}

func ConvertPeersStats(ss []*comm.PeerStats) []*PeerStats {
	if len(ss) == 0 {
		return nil
	}
	peersStats := make([]*PeerStats, len(ss))
	for i, peerStats := range ss {
		peersStats[i] = &PeerStats{
			Name:        peerStats.Name,
			BestBlockID: peerStats.BestBlockID,
			TotalScore:  peerStats.TotalScore,
			PeerID:      peerStats.PeerID,
			NetAddr:     peerStats.NetAddr,
			Inbound:     peerStats.Inbound,
			Duration:    peerStats.Duration,
			PeerAddress: peerStats.PeerAddress,
		}
	}
	return peersStats
}
