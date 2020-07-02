// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package comm

import (
	"github.com/HiNounou029/nounouchain/polo"
)

// type Traffic struct {
// 	Bytes    uint64
// 	Requests uint64
// 	Errors   uint64
// }

// PeerStats records stats of a peer.
type PeerStats struct {
	Name        string
	BestBlockID polo.Bytes32
	TotalScore  uint64
	PeerID      string
	NetAddr     string
	Inbound     bool
	Duration    uint64 // in seconds
	PeerAddress string
}
