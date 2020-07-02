// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package xenv

import (
	"math/big"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/abi"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/vm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
)

// BlockContext block context.
type BlockContext struct {
	Beneficiary polo.Address
	Signer      polo.Address
	Number      uint32
	Time        uint64
	GasLimit    uint64
	TotalScore  uint64
}

// TransactionContext transaction context.
type TransactionContext struct {
	ID         polo.Bytes32
	Origin     polo.Address
	GasPrice   *big.Int
	BlockRef   tx.BlockRef
	Expiration uint32
}

// Environment an env to execute native method.
type Environment struct {
	abi      *abi.Method
	seeker   *chain.Seeker
	state    *state.State
	blockCtx *BlockContext
	txCtx    *TransactionContext
	evm      *vm.EVM
	contract *vm.Contract
}

// New create a new env.
func New(
	abi *abi.Method,
	seeker *chain.Seeker,
	state *state.State,
	blockCtx *BlockContext,
	txCtx *TransactionContext,
	evm *vm.EVM,
	contract *vm.Contract,
) *Environment {
	return &Environment{
		abi:      abi,
		seeker:   seeker,
		state:    state,
		blockCtx: blockCtx,
		txCtx:    txCtx,
		evm:      evm,
		contract: contract,
	}
}

func (env *Environment) Seeker() *chain.Seeker                   { return env.seeker }
func (env *Environment) State() *state.State                     { return env.state }
func (env *Environment) TransactionContext() *TransactionContext { return env.txCtx }
func (env *Environment) BlockContext() *BlockContext             { return env.blockCtx }
func (env *Environment) Caller() polo.Address                    { return polo.Address(env.contract.Caller()) }
func (env *Environment) To() polo.Address                        { return polo.Address(env.contract.Address()) }

func (env *Environment) UseGas(gas uint64) {
	if !env.contract.UseGas(gas) {
		panic(vm.ErrOutOfGas)
	}
}

func (env *Environment) ParseArgs(val interface{}) {
	if err := env.abi.DecodeInput(env.contract.Input, val); err != nil {
		// as vm error
		panic(errors.WithMessage(err, "decode native input"))
	}
}

func (env *Environment) Log(abi *abi.Event, address polo.Address, topics []polo.Bytes32, args ...interface{}) {
	data, err := abi.Encode(args...)
	if err != nil {
		panic(errors.WithMessage(err, "encode native event"))
	}
	env.UseGas(ethparams.LogGas + ethparams.LogTopicGas*uint64(len(topics)) + ethparams.LogDataGas*uint64(len(data)))

	ethTopics := make([]common.Hash, 0, len(topics)+1)
	ethTopics = append(ethTopics, common.Hash(abi.ID()))
	for _, t := range topics {
		ethTopics = append(ethTopics, common.Hash(t))
	}
	env.evm.StateDB.AddLog(&types.Log{
		Address: common.Address(address),
		Topics:  ethTopics,
		Data:    data,
	})
}

func (env *Environment) Call(proc func(env *Environment) []interface{}) (output []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			if e == vm.ErrOutOfGas {
				err = vm.ErrOutOfGas
			} else {
				panic(e)
			}
		}
	}()
	data, err := env.abi.EncodeOutput(proc(env)...)
	if err != nil {
		panic(errors.WithMessage(err, "encode native output"))
	}
	return data, nil
}
