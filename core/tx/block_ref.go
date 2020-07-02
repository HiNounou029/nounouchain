// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package tx

import (
	"encoding/binary"
	"github.com/HiNounou029/nounouchain/polo"
)

// BlockRef is block reference.
type BlockRef [8]byte

// Number extracts block number.
func (br BlockRef) Number() uint32 {
	//	number := binary.BigEndian.Uint32(br[:])
	//	str := fmt.Sprint(number)
	//	fmt.Println("block number: "+str)
	return binary.BigEndian.Uint32(br[:])
}

// NewBlockRef create block reference with block number.
func NewBlockRef(blockNum uint32) (br BlockRef) {
	binary.BigEndian.PutUint32(br[:], blockNum)
	return
}

// NewBlockRefFromID create block reference from block id.
func NewBlockRefFromID(blockID polo.Bytes32) (br BlockRef) {
	copy(br[:], blockID[:])
	return
}
