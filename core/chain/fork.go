// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package chain

import (
	"github.com/HiNounou029/nounouchain/core/block"
)

// Fork describes forked chain.
type Fork struct {
	Ancestor *block.Header
	Trunk    []*block.Header
	Branch   []*block.Header
}
