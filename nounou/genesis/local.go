// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package genesis

import (
	"crypto/ecdsa"
	"math/big"
	"sync/atomic"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/builtin"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/vm"
	"github.com/HiNounou029/nounouchain/crypto"
)

// DevAccount account for development.
type DevAccount struct {
	Address    polo.Address
	PrivateKey *ecdsa.PrivateKey
}

var devAccounts atomic.Value

// DevAccounts returns pre-alloced accounts for solo mode.
func DevAccounts() []DevAccount {
	if accs := devAccounts.Load(); accs != nil {
		return accs.([]DevAccount)
	}

	var accs []DevAccount
	privKeys := []string{
		"dce1443bd2ef0c2631adc1c67e5c93f13dc23a41c18b536effbbdcbcdb96fb65",
		"321d6443bc6177273b5abf54210fe806d451d6b7973bccc2384ef78bbcd0bf51",
		"2d7c882bad2a01105e36dda3646693bc1aaaa45b0ed63fb0ce23c060294f3af2",
		"593537225b037191d322c3b1df585fb1e5100811b71a6f7fc7e29cca1333483e",
		"ca7b25fc980c759df5f3ce17a3d881d6e19a38e651fc4315fc08917edab41058",
		"88d2d80b12b92feaa0da6d62309463d20408157723f2d7e799b6a74ead9a673b",
		"fbb9e7ba5fe9969a71c6599052237b91adeb1e5fc0c96727b66e56ff5d02f9d0",
		"547fb081e73dc2e22b4aae5c60e2970b008ac4fc3073aebc27d41ace9c4f53e9",
		"c8c53657e41a8d669349fc287f57457bd746cb1fcfc38cf94d235deb2cfca81b",
		"87e0eba9c86c494d98353800571089f316740b0cb84c9a7cdf2fe5c9997c7966",
	}
	for _, str := range privKeys {
		pk, err := crypto.HexToECDSA(str)
		if err != nil {
			panic(err)
		}
		addr := crypto.PubkeyToAddress(pk.PublicKey)
		accs = append(accs, DevAccount{polo.Address(addr), pk})
	}
	devAccounts.Store(accs)
	return accs
}

// NewDevnet create genesis for solo mode.
func NewDevnet() *Genesis {
	launchTime := uint64(1542816000) // '2018-11-22 00:00:00 +0800 CST'

	executor := DevAccounts()[0].Address
	soloBlockSigner := DevAccounts()[0]

	builder := new(Builder).
		GasLimit(polo.InitialGasLimit).
		Timestamp(launchTime).
		State(func(state *state.State) error {
			// alloc precompiled contracts
			for addr := range vm.PrecompiledContractsByzantium {
				state.SetCode(polo.Address(addr), emptyRuntimeBytecode)
			}

			// setup builtin contracts
			state.SetCode(builtin.Authority.Address, builtin.Authority.RuntimeBytecodes())
			state.SetCode(builtin.Params.Address, builtin.Params.RuntimeBytecodes())
			state.SetCode(builtin.Prototype.Address, builtin.Prototype.RuntimeBytecodes())
			state.SetCode(builtin.Extension.Address, builtin.Extension.RuntimeBytecodes())

			for _, a := range DevAccounts() {
				bal, _ := new(big.Int).SetString("1000000000000000000000000000", 10)
				state.SetBalance(a.Address, bal)
			}
			return nil
		}).
		Call(
			tx.NewClause(&builtin.Params.Address).WithData(mustEncodeInput(builtin.Params.ABI, "set", polo.KeyExecutorAddress, new(big.Int).SetBytes(executor[:]))),
			polo.Address{}).
		Call(
			tx.NewClause(&builtin.Params.Address).WithData(mustEncodeInput(builtin.Params.ABI, "set", polo.KeyBaseGasPrice, polo.InitialBaseGasPrice)),
			executor).
		Call(
			tx.NewClause(&builtin.Params.Address).WithData(mustEncodeInput(builtin.Params.ABI, "set", polo.KeyProposerEndorsement, polo.InitialProposerEndorsement)),
			executor).
		Call(
			tx.NewClause(&builtin.Authority.Address).WithData(mustEncodeInput(builtin.Authority.ABI, "add", soloBlockSigner.Address, soloBlockSigner.Address, polo.BytesToBytes32([]byte("Solo Block Signer")))),
			executor)

	id, err := builder.ComputeID()
	if err != nil {
		panic(err)
	}

	return &Genesis{builder, id, "devnet"}
}
