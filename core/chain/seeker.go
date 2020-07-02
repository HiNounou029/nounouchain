// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package chain

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/block"
)

// Seeker to seek block by given number on the chain defined by head block ID.
type Seeker struct {
	chain       *Chain
	headBlockID polo.Bytes32
	err         error
}

func newSeeker(chain *Chain, headBlockID polo.Bytes32) *Seeker {
	return &Seeker{
		chain:       chain,
		headBlockID: headBlockID,
	}
}

func (s *Seeker) setError(err error) {
	if s.err == nil {
		s.err = err
	}
}

// Err returns error occurred.
func (s *Seeker) Err() error {
	return s.err
}

// GetID returns block ID by the given number.
func (s *Seeker) GetID(num uint32) polo.Bytes32 {
	if num > block.Number(s.headBlockID) {
		panic("num exceeds head block")
	}
	id, err := s.chain.GetAncestorBlockID(s.headBlockID, num)
	s.setError(err)
	return id
}

// GetHeader returns block header by the given number.
func (s *Seeker) GetHeader(id polo.Bytes32) *block.Header {
	header, err := s.chain.GetBlockHeader(id)
	if err != nil {
		s.setError(err)
		return &block.Header{}
	}
	return header
}

// GenesisID get genesis block ID.
func (s *Seeker) GenesisID() polo.Bytes32 {
	return s.chain.GenesisBlock().Header().ID()
}
