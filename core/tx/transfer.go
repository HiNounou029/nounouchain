// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package tx

import (
	"math/big"

	"github.com/HiNounou029/nounouchain/polo"
)

// Transfer token transfer log.
type Transfer struct {
	Sender    polo.Address
	Recipient polo.Address
	Amount    *big.Int
}

// Transfers slice of transfer logs.
type Transfers []*Transfer
