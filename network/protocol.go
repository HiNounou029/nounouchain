// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package network

import "github.com/ethereum/go-ethereum/p2p"

// Protocol represents a P2P subprotocol implementation.
type Protocol struct {
	p2p.Protocol

	DiscTopic string
}
