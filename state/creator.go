// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package state

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/storage/kv"
)

// Creator state creator to cut-off kv dependency.
type Creator struct {
	kv kv.GetPutter
}

// NewCreator create a new state creator.
func NewCreator(kv kv.GetPutter) *Creator {
	return &Creator{kv}
}

// NewState create a new state object.
func (c *Creator) NewState(root polo.Bytes32) (*State, error) {
	return New(root, c.kv)
}
