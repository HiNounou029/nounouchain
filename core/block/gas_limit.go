// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package block

import (
	"github.com/HiNounou029/nounouchain/polo"
)

// GasLimit to support block gas limit validation and adjustment.
type GasLimit uint64

// IsValid returns if the receiver is valid according to parent gas limit.
func (gl GasLimit) IsValid(parentGasLimit uint64) bool {
	gasLimit := uint64(gl)
	if gasLimit < polo.MinGasLimit {
		return false
	}
	var diff uint64
	if gasLimit > parentGasLimit {
		diff = gasLimit - parentGasLimit
	} else {
		diff = parentGasLimit - gasLimit
	}

	return diff <= parentGasLimit/polo.GasLimitBoundDivisor
}