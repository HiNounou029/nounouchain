// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package builtin_test

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/builtin"
	"github.com/HiNounou029/nounouchain/common/xenv"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/HiNounou029/nounouchain/vm/runtime"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/stretchr/testify/assert"
)

func approverEvent(approver polo.Address, action string) *tx.Event {
	ev, _ := builtin.Executor.ABI.EventByName("Approver")
	var b32 polo.Bytes32
	copy(b32[:], action)
	data, _ := ev.Encode(b32)
	return &tx.Event{
		Address: builtin.Executor.Address,
		Topics:  []polo.Bytes32{ev.ID(), polo.BytesToBytes32(approver.Bytes())},
		Data:    data,
	}
}
func proposalEvent(id polo.Bytes32, action string) *tx.Event {
	ev, _ := builtin.Executor.ABI.EventByName("Proposal")
	var b32 polo.Bytes32
	copy(b32[:], action)
	data, _ := ev.Encode(b32)
	return &tx.Event{
		Address: builtin.Executor.Address,
		Topics:  []polo.Bytes32{ev.ID(), id},
		Data:    data,
	}
}

func votingContractEvent(addr polo.Address, action string) *tx.Event {
	ev, _ := builtin.Executor.ABI.EventByName("VotingContract")
	var b32 polo.Bytes32
	copy(b32[:], action)
	data, _ := ev.Encode(b32)
	return &tx.Event{
		Address: builtin.Executor.Address,
		Topics:  []polo.Bytes32{ev.ID(), polo.BytesToBytes32(addr.Bytes())},
		Data:    data,
	}
}

func initExectorTest() *ctest {
	kv, _ := storage.NewMem()
	b0 := buildGenesis(kv, func(state *state.State) error {
		state.SetCode(builtin.Prototype.Address, builtin.Prototype.RuntimeBytecodes())
		state.SetCode(builtin.Executor.Address, builtin.Executor.RuntimeBytecodes())
		state.SetCode(builtin.Params.Address, builtin.Params.RuntimeBytecodes())
		builtin.Params.Native(state).Set(polo.KeyExecutorAddress, new(big.Int).SetBytes(builtin.Executor.Address[:]))
		return nil
	})

	c, _ := chain.New(kv, b0)
	st, _ := state.New(b0.Header().StateRoot(), kv)
	seeker := c.NewSeeker(b0.Header().ID())

	rt := runtime.New(seeker, st, &xenv.BlockContext{Time: uint64(time.Now().Unix())})

	return &ctest{
		rt:  rt,
		abi: builtin.Executor.ABI,
		to:  builtin.Executor.Address,
	}
}

func TestExecutorApprover(t *testing.T) {
	test := initExectorTest()
	var approvers []polo.Address
	for i := 0; i < 7; i++ {
		approvers = append(approvers, polo.BytesToAddress([]byte(fmt.Sprintf("approver%d", i))))
	}

	for _, a := range approvers {
		// zero identity
		test.Case("addApprover", a, polo.Bytes32{}).
			ShouldVMError(errReverted).
			Assert(t)

		test.Case("addApprover", a, polo.BytesToBytes32(a.Bytes())).
			Caller(polo.BytesToAddress([]byte("other"))).
			ShouldVMError(errReverted).
			Assert(t)

		test.Case("addApprover", a, polo.BytesToBytes32(a.Bytes())).
			Caller(builtin.Executor.Address).
			ShouldLog(approverEvent(a, "added")).
			Assert(t)
		assert.True(t, builtin.Prototype.Native(test.rt.State()).Bind(test.to).IsUser(a))
	}

	test.Case("approverCount").
		ShouldOutput(uint8(len(approvers))).
		Assert(t)

	test.Case("addApprover", approvers[0], polo.BytesToBytes32(approvers[0].Bytes())).
		Caller(builtin.Executor.Address).
		ShouldVMError(errReverted).
		Assert(t)

	for _, a := range approvers {
		test.Case("approvers", a).
			ShouldOutput(polo.BytesToBytes32(a.Bytes()), true).
			Assert(t)
	}

	for _, a := range approvers {
		test.Case("revokeApprover", a).
			ShouldVMError(errReverted).
			Assert(t)

		test.Case("revokeApprover", a).
			Caller(builtin.Executor.Address).
			ShouldLog(approverEvent(a, "revoked")).
			Assert(t)
		assert.False(t, builtin.Prototype.Native(test.rt.State()).Bind(test.to).IsUser(a))
	}
	test.Case("approverCount").
		ShouldOutput(uint8(0)).
		Assert(t)
}

func TestExecutorVotingContract(t *testing.T) {

	test := initExectorTest()
	voting := polo.BytesToAddress([]byte("voting"))
	test.Case("attachVotingContract", voting).
		ShouldVMError(errReverted).
		Assert(t)
	test.Case("votingContracts", voting).
		ShouldOutput(false).
		Assert(t)
	test.Case("attachVotingContract", voting).
		Caller(builtin.Executor.Address).
		ShouldLog(votingContractEvent(voting, "attached")).
		Assert(t)

	test.Case("votingContracts", voting).
		ShouldOutput(true).
		Assert(t)

	test.Case("attachVotingContract", voting).
		Caller(builtin.Executor.Address).
		ShouldVMError(errReverted).
		Assert(t)

	test.Case("detachVotingContract", voting).
		Caller(builtin.Executor.Address).
		ShouldLog(votingContractEvent(voting, "detached")).
		Assert(t)

	test.Case("attachVotingContract", voting).
		Caller(builtin.Executor.Address).
		ShouldLog(votingContractEvent(voting, "attached")).
		Assert(t)
}

func TestExecutorProposal(t *testing.T) {
	test := initExectorTest()

	target := builtin.Params.Address
	setParam, _ := builtin.Params.ABI.MethodByName("set")
	data, _ := setParam.EncodeInput(polo.BytesToBytes32([]byte("paramKey")), big.NewInt(12345))
	test.Case("propose", target, data).
		ShouldVMError(errReverted).
		Assert(t)

	approver := polo.BytesToAddress([]byte("approver"))
	test.Case("addApprover", approver, polo.BytesToBytes32(approver.Bytes())).
		Caller(builtin.Executor.Address).
		Assert(t)

	proposalID := func() polo.Bytes32 {
		var b8 [8]byte
		binary.BigEndian.PutUint64(b8[:], test.rt.Context().Time)
		return polo.Bytes32(crypto.Keccak256Hash(b8[:], approver[:]))
	}()
	test.Case("propose", target, data).
		Caller(approver).
		ShouldOutput(proposalID).
		ShouldLog(proposalEvent(proposalID, "proposed")).
		Assert(t)

	test.Case("proposals", proposalID).
		ShouldOutput(
			test.rt.Context().Time,
			approver,
			uint8(1),
			uint8(0),
			false,
			target,
			data).
		Assert(t)

	test.Case("approve", proposalID).
		ShouldVMError(errReverted).
		Assert(t)

	test.Case("execute", proposalID).
		ShouldVMError(errReverted).
		Assert(t)

	test.Case("approve", proposalID).
		Caller(approver).
		ShouldLog(proposalEvent(proposalID, "approved")).
		Assert(t)
	test.Case("proposals", proposalID).
		ShouldOutput(
			test.rt.Context().Time,
			approver,
			uint8(1),
			uint8(1),
			false,
			target,
			data).
		Assert(t)

	test.Case("execute", proposalID).
		ShouldLog(proposalEvent(proposalID, "executed")).
		Assert(t)

	test.Case("execute", proposalID).
		ShouldVMError(errReverted).
		Assert(t)
	test.Case("proposals", proposalID).
		ShouldOutput(
			test.rt.Context().Time,
			approver,
			uint8(1),
			uint8(1),
			true,
			target,
			data).
		Assert(t)

	assert.Equal(t, builtin.Params.Native(test.rt.State()).Get(polo.BytesToBytes32([]byte("paramKey"))), big.NewInt(12345))
}
