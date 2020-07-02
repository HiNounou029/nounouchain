// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

// +build gm

package crypto

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/tjfoc/gmsm/sm3"
)

// Keccak256 calculates and returns the sm3 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	d := sm3.New()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// Keccak256Hash calculates and returns the sm3 hash of the input data,
// converting it to an internal Hash data structure.
func Keccak256Hash(data ...[]byte) (h common.Hash) {
	d := sm3.New()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}
