// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package rpc

import "github.com/ethereum/go-ethereum/p2p"

type msgData struct {
	ID       uint32
	IsResult bool
	Payload  interface{}
}

type resultListener struct {
	msgCode  uint64
	onResult func(*p2p.Msg) error
}
