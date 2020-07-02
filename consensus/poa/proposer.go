// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package poa

import (
	"github.com/HiNounou029/nounouchain/polo"
)

// Proposer address with status.
type Proposer struct {
	Address polo.Address
	Active  bool
}
