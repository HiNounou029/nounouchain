// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

// +build gm

package sha3

import (
	"github.com/tjfoc/gmsm/sm3"
	"hash"
)

// NewKeccak256 creates a new sm3 hash.
func NewKeccak256() hash.Hash {
	return sm3.New()
}
