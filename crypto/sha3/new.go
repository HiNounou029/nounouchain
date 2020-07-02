// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

// +build !gm

package sha3

import (
	"golang.org/x/crypto/sha3"
	"hash"
)

// NewKeccak256 creates a new sha3-256 hash.
func NewKeccak256() hash.Hash { return sha3.New256() }
