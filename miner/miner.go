// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package miner

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/builtin"
	"github.com/HiNounou029/nounouchain/common/xenv"
	"github.com/HiNounou029/nounouchain/consensus/poa"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/vm/runtime"
	"github.com/pkg/errors"
	"math/big"
)

// Miner to pack txs and build new blocks.
type Miner struct {
	chain          *chain.Chain
	stateCreator   *state.Creator
	nodeMaster     polo.Address
	beneficiary    *polo.Address
	targetGasLimit uint64
}

// New create a new Miner instance.
// The beneficiary is optional, it defaults to endorsor if not set.
func New(
	chain *chain.Chain,
	stateCreator *state.Creator,
	nodeMaster polo.Address,
	beneficiary *polo.Address) *Miner {

	return &Miner{
		chain,
		stateCreator,
		nodeMaster,
		beneficiary,
		0,
	}
}

// Schedule schedule a packing flow to pack new block upon given parent and clock time.
func (p *Miner) Schedule(parent *block.Header, nowTimestamp uint64) (flow *Flow, err error) {
	state, err := p.stateCreator.NewState(parent.StateRoot())
	if err != nil {
		return nil, errors.Wrap(err, "state")
	}

	var (
//		endorsement = builtin.Params.Native(state).Get(polo.KeyProposerEndorsement)
		endorsement = big.NewInt(0)
		authority   = builtin.Authority.Native(state)
		candidates  = authority.Candidates(endorsement, polo.Conf.MaxBlockProposers)
		proposers   = make([]poa.Proposer, 0, len(candidates))
		beneficiary polo.Address
	)
	if p.beneficiary != nil {
		beneficiary = *p.beneficiary
	}

	for _, c := range candidates {
		if p.beneficiary == nil && c.NodeMaster == p.nodeMaster {
			// not beneficiary not set, set it to endorsor
			beneficiary = c.Endorsor
		}
		proposers = append(proposers, poa.Proposer{
			Address: c.NodeMaster,
			Active:  c.Active,
		})
	}

	// calc the time when it's turn to produce block
	sched, err := poa.NewScheduler(p.nodeMaster, proposers, parent.Number(), parent.Timestamp())
	if err != nil {
		return nil, err
	}

	newBlockTime := sched.Schedule(nowTimestamp)
	updates, score := sched.Updates(newBlockTime)

	for _, u := range updates {
		authority.Update(u.Address, u.Active)
	}

	rt := runtime.New(
		p.chain.NewSeeker(parent.ID()),
		state,
		&xenv.BlockContext{
			Beneficiary: beneficiary,
			Signer:      p.nodeMaster,
			Number:      parent.Number() + 1,
			Time:        newBlockTime,
			GasLimit:    p.gasLimit(parent.GasLimit()),
			TotalScore:  parent.TotalScore() + score,
		})

	return newFlow(p, parent, rt), nil
}

func (p *Miner) gasLimit(parentGasLimit uint64) uint64 {
	if p.targetGasLimit > 0 {
		return p.targetGasLimit
	}
	return parentGasLimit
}

// SetTargetGasLimit set target gas limit, the Miner will adjust block gas limit close to
// it as it can.
func (p *Miner) SetTargetGasLimit(gl uint64) {
	p.targetGasLimit = gl
}
