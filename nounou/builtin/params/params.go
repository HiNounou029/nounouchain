// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package params

import (
	"math/big"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/ethereum/go-ethereum/rlp"
)

// Params binder of `Params` contract.
type Params struct {
	addr  polo.Address
	state *state.State
}

func New(addr polo.Address, state *state.State) *Params {
	return &Params{addr, state}
}

// Get native way to get param.
func (p *Params) Get(key polo.Bytes32) (value *big.Int) {
	p.state.DecodeStorage(p.addr, key, func(raw []byte) error {
		if len(raw) == 0 {
			value = &big.Int{}
			return nil
		}
		return rlp.DecodeBytes(raw, &value)
	})
	return
}

// Set native way to set param.
func (p *Params) Set(key polo.Bytes32, value *big.Int) {
	p.state.EncodeStorage(p.addr, key, func() ([]byte, error) {
		if value.Sign() == 0 {
			return nil, nil
		}
		return rlp.EncodeToBytes(value)
	})
}
