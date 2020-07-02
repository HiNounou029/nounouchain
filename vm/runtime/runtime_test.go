// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package runtime_test

import (
	"encoding/hex"
	"math"
	"math/big"
	"testing"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/abi"
	"github.com/HiNounou029/nounouchain/nounou/builtin"
	"github.com/HiNounou029/nounouchain/nounou/genesis"
	"github.com/HiNounou029/nounouchain/common/xenv"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/HiNounou029/nounouchain/vm/runtime"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestContractSuicide(t *testing.T) {
	assert := assert.New(t)
	kv, _ := storage.NewMem()

	g := genesis.NewDevnet()
	stateCreator := state.NewCreator(kv)
	b0, _, err := g.Build(stateCreator)
	if err != nil {
		t.Fatal(err)
	}
	ch, _ := chain.New(kv, b0)

	// contract:
	//
	// pragma solidity ^0.4.18;

	// contract TestSuicide {
	// 	function testSuicide() public {
	// 		selfdestruct(msg.sender);
	// 	}
	// }
	data, _ := hex.DecodeString("608060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680639799f6ba146044575b600080fd5b348015604f57600080fd5b5060566058565b005b3373ffffffffffffffffffffffffffffffffffffffff16ff00a165627a7a72305820edbdbbe65c66633c944b4492bec73d7226a2184829f0a3677ccf5eab67d9fc060029")
	time := b0.Header().Timestamp()
	addr := polo.BytesToAddress([]byte("acc01"))
	state, _ := stateCreator.NewState(b0.Header().StateRoot())
	state.SetCode(addr, data)
	state.SetBalance(addr, big.NewInt(200))

	abi, _ := abi.New([]byte(`[{
	"constant": false,
	"inputs": [],
	"name": "testSuicide",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}]`))
	suicide, _ := abi.MethodByName("testSuicide")
	methodData, err := suicide.EncodeInput()
	if err != nil {
		t.Fatal(err)
	}

	origin := genesis.DevAccounts()[0].Address
	out := runtime.New(ch.NewSeeker(b0.Header().ID()), state, &xenv.BlockContext{Time: time}).
		ExecuteClause(tx.NewClause(&addr).WithData(methodData), 0, math.MaxUint64, &xenv.TransactionContext{Origin: origin})
	if out.VMErr != nil {
		t.Fatal(out.VMErr)
	}

	expectedTransfer := &tx.Transfer{
		Sender:    addr,
		Recipient: origin,
		Amount:    big.NewInt(200),
	}
	assert.Equal(1, len(out.Transfers))
	assert.Equal(expectedTransfer, out.Transfers[0])

	assert.Equal(0, len(out.Events))

	assert.Equal(big.NewInt(0).String(), state.GetBalance(addr).String())

	bal, _ := new(big.Int).SetString("1000000000000000000000000000", 10)
	assert.Equal(new(big.Int).Add(bal, big.NewInt(200)), state.GetBalance(origin))
}

func TestCall(t *testing.T) {
	kv, _ := storage.NewMem()

	g := genesis.NewDevnet()
	b0, _, err := g.Build(state.NewCreator(kv))
	if err != nil {
		t.Fatal(err)
	}

	ch, _ := chain.New(kv, b0)

	state, _ := state.New(b0.Header().StateRoot(), kv)

	rt := runtime.New(ch.NewSeeker(b0.Header().ID()), state, &xenv.BlockContext{})

	method, _ := builtin.Params.ABI.MethodByName("executor")
	data, err := method.EncodeInput()
	if err != nil {
		t.Fatal(err)
	}

	out := rt.ExecuteClause(
		tx.NewClause(&builtin.Params.Address).WithData(data),
		0, math.MaxUint64, &xenv.TransactionContext{})

	if out.VMErr != nil {
		t.Fatal(out.VMErr)
	}

	var addr common.Address
	if err := method.DecodeOutput(out.Data, &addr); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, polo.Address(addr), genesis.DevAccounts()[0].Address)

	data, _ = hex.DecodeString("6080604052348015600f57600080fd5b505b600115601b576011565b60358060286000396000f3006080604052600080fd00a165627a7a7230582026c386600e61384b3a93bf45760f3207b5cac072cec31c9cea1bc7099bda49b00029")
	exec, interrupt := rt.PrepareClause(tx.NewClause(nil).WithData(data), 0, math.MaxUint64, &xenv.TransactionContext{})

	go func() {
		interrupt()
	}()

	out, interrupted := exec()
	assert.NotNil(t, out)
	assert.True(t, interrupted)
}
