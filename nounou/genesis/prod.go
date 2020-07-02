// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package genesis

import (
	"math/big"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/builtin"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/vm"
)

// NewProdnet create cefanet genesis.
func NewProdnet(cfgFilePath string) *Genesis {
	launchTime := uint64(1542816000) // '2018-11-22 00:00:00 +0800 CST'

	genesisCfg := MustReadConfig(cfgFilePath)

	initialAuthorityNodes := loadAuthorityNodes(genesisCfg)
	approvers := loadApprovers(genesisCfg)

	builder := new(Builder).
		Timestamp(launchTime).
		GasLimit(polo.InitialGasLimit).
		State(func(state *state.State) error {
			// alloc precompiled contracts
			for addr := range vm.PrecompiledContractsByzantium {
				state.SetCode(polo.Address(addr), emptyRuntimeBytecode)
			}

			// alloc builtin contracts
			state.SetCode(builtin.Authority.Address, builtin.Authority.RuntimeBytecodes())
			state.SetCode(builtin.Executor.Address, builtin.Executor.RuntimeBytecodes())
			state.SetCode(builtin.Extension.Address, builtin.Extension.RuntimeBytecodes())
			state.SetCode(builtin.Params.Address, builtin.Params.RuntimeBytecodes())
			state.SetCode(builtin.Prototype.Address, builtin.Prototype.RuntimeBytecodes())

			// alloc tokens for authority node endorsor
			for _, anode := range initialAuthorityNodes {
				state.SetBalance(anode.endorsorAddress, polo.InitialProposerEndorsement)
			}

			// alloc tokens for approvers
			amount := new(big.Int).Mul(big.NewInt(210469086165), big.NewInt(1e17))
			for _, approver := range approvers {
				state.SetBalance(approver.address, amount)
			}

			return nil
		})

	///// initialize builtin contracts

	// initialize params
	data := mustEncodeInput(builtin.Params.ABI, "set", polo.KeyExecutorAddress, new(big.Int).SetBytes(builtin.Executor.Address[:]))
	builder.Call(tx.NewClause(&builtin.Params.Address).WithData(data), polo.Address{})

	data = mustEncodeInput(builtin.Params.ABI, "set", polo.KeyBaseGasPrice, polo.InitialBaseGasPrice)
	builder.Call(tx.NewClause(&builtin.Params.Address).WithData(data), builtin.Executor.Address)

	data = mustEncodeInput(builtin.Params.ABI, "set", polo.KeyProposerEndorsement, polo.InitialProposerEndorsement)
	builder.Call(tx.NewClause(&builtin.Params.Address).WithData(data), builtin.Executor.Address)

	// add initial authority nodes
	for _, anode := range initialAuthorityNodes {
		data := mustEncodeInput(builtin.Authority.ABI, "add", anode.masterAddress, anode.endorsorAddress, anode.identity)
		builder.Call(tx.NewClause(&builtin.Authority.Address).WithData(data), builtin.Executor.Address)
	}

	// add initial approvers (steering committee)
	for _, approver := range approvers {
		data := mustEncodeInput(builtin.Executor.ABI, "addApprover", approver.address, polo.BytesToBytes32([]byte(approver.identity)))
		builder.Call(tx.NewClause(&builtin.Executor.Address).WithData(data), builtin.Executor.Address)
	}

	var extra [28]byte
	copy(extra[:], "CefaChain")
	builder.ExtraData(extra)
	id, err := builder.ComputeID()
	if err != nil {
		panic(err)
	}
	return &Genesis{builder, id, "cefanet"}
}

type authorityNode struct {
	masterAddress   polo.Address
	endorsorAddress polo.Address
	identity        polo.Bytes32
}

type approver struct {
	address  polo.Address
	identity string
}

func loadApprovers(cfg *Config) []*approver {
	approvers := make([]*approver, len(cfg.Approvers))
	for i, account := range cfg.Approvers {
		approvers[i] = &approver{
			address:  polo.MustParseAddress(account.Address),
			identity: account.Id,
		}
	}
	return approvers
}

func loadAuthorityNodes(cfg *Config) []*authorityNode {
	candidates := make([]*authorityNode, len(cfg.Authorities))
	for i, account := range cfg.Authorities {
		candidates[i] = &authorityNode{
			masterAddress:   polo.MustParseAddress(account.Address),
			endorsorAddress: polo.MustParseAddress(account.Address),
			identity:        polo.MustParseBytes32(account.Id),
		}
	}
	return candidates
}
