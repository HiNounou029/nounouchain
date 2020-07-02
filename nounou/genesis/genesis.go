// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package genesis

import (
	"encoding/hex"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/abi"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
)

// Genesis to build genesis block.
type Genesis struct {
	builder *Builder
	id      polo.Bytes32
	name    string
}

// Build build the genesis block.
func (g *Genesis) Build(stateCreator *state.Creator) (blk *block.Block, events tx.Events, err error) {
	block, events, err := g.builder.Build(stateCreator)
	if err != nil {
		return nil, nil, err
	}
	if block.Header().ID() != g.id {
		panic("built genesis ID incorrect")
	}
	return block, events, nil
}

// ID returns genesis block ID.
func (g *Genesis) ID() polo.Bytes32 {
	return g.id
}

// Name returns network name.
func (g *Genesis) Name() string {
	return g.name
}

func MustEncodeInput(abi *abi.ABI, name string, args ...interface{}) []byte {
	return mustEncodeInput(abi, name, args)
}

func mustEncodeInput(abi *abi.ABI, name string, args ...interface{}) []byte {
	m, found := abi.MethodByName(name)
	if !found {
		panic("method not found")
	}
	data, err := m.EncodeInput(args...)
	if err != nil {
		panic(err)
	}
	return data
}

func mustDecodeHex(str string) []byte {
	data, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return data
}

var emptyRuntimeBytecode = mustDecodeHex("6060604052600256")
