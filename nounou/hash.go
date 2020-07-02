// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package polo

import (
	"github.com/HiNounou029/nounouchain/crypto/sha3"
	"hash"
)

func NewBlake2b() hash.Hash {
	hash := sha3.NewKeccak256()
	return hash
}

// Blake2b computes blake2b-256 checksum for given data.
func Blake2b(data ...[]byte) (b32 Bytes32) {
	hash := NewBlake2b()
	for _, b := range data {
		hash.Write(b)
	}
	hash.Sum(b32[:0])
	return
}

/*
// NewBlake2b return blake2b-256 hash.
func NewBlake2b() hash.Hash {
	hash := sha3.NewKeccak256()
	return hash
}

// Blake2b computes blake2b-256 checksum for given data.
func Blake2b(data ...[]byte) (b32 Bytes32) {
	hash := sha3.NewKeccak256()
	for _, b := range data {
		hash.Write(b)
	}
	hash.Sum(b32[:0])
	return
}
*/
