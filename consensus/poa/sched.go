// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package poa

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"

	"github.com/HiNounou029/nounouchain/polo"
)

// Scheduler to schedule the time when a proposer to produce a block.
type Scheduler struct {
	proposer          Proposer
	actives           []Proposer
	parentBlockNumber uint32
	parentBlockTime   uint64
}

// NewScheduler create a Scheduler object.
// `addr` is the proposer to be scheduled.
// If `addr` is not listed in `proposers`, an error returned.
func NewScheduler(
	addr polo.Address,
	proposers []Proposer,
	parentBlockNumber uint32,
	parentBlockTime uint64) (*Scheduler, error) {

	actives := make([]Proposer, 0, len(proposers))
	listed := false
	var proposer Proposer
	for _, p := range proposers {
		if p.Address == addr {
			proposer = p
			actives = append(actives, p)
			listed = true
		} else if p.Active {
			actives = append(actives, p)
		}
	}

	if !listed {
		return nil, errors.New("unauthorized block proposer")
	}

	return &Scheduler{
		proposer,
		actives,
		parentBlockNumber,
		parentBlockTime,
	}, nil
}

func (s *Scheduler) whoseTurn(t uint64) Proposer {
	index := dprp(s.parentBlockNumber, t) % uint64(len(s.actives))
	return s.actives[index]
}

var singleCount = 0

// Schedule to determine time of the proposer to produce a block, according to `nowTime`.
// `newBlockTime` is promised to be >= nowTime and > parentBlockTime
func (s *Scheduler) Schedule(nowTime uint64) (newBlockTime uint64) {
	var T = polo.Conf.BlockInterval

	if s.parentBlockTime%T == 0 {
		newBlockTime = s.parentBlockTime + T
	} else {
		newBlockTime = (s.parentBlockTime/T)*T + T + T
	}

	if nowTime > newBlockTime {
		// ensure T aligned, and >= nowTime
		newBlockTime += (nowTime - newBlockTime + T - 1) / T * T
	}

	if len(s.actives) == 1 {
		singleCount++
	}
	if singleCount >= 10 {
		r := uint64(rand.Intn(int(T * 4)))
		r = r / T * T
		fmt.Println("add random time:", r)
		newBlockTime += uint64(r)
		singleCount = 0
	}

	for {
		p := s.whoseTurn(newBlockTime)
		if p.Address == s.proposer.Address {
			return newBlockTime
		}

		// try next time slot
		newBlockTime += T
	}
}

// IsTheTime returns if the newBlockTime is correct for the proposer.
func (s *Scheduler) IsTheTime(newBlockTime uint64) bool {
	if s.parentBlockTime >= newBlockTime {
		// invalid block time
		return false
	}

	if (newBlockTime-s.parentBlockTime)%polo.Conf.BlockInterval != 0 {
		// invalid block time
		return false
	}

	return s.whoseTurn(newBlockTime).Address == s.proposer.Address
}

// Updates returns proposers whose status are change, and the score when new block time is assumed to be newBlockTime.
func (s *Scheduler) Updates(newBlockTime uint64) (updates []Proposer, score uint64) {

	toDeactivate := make(map[polo.Address]Proposer)

	t := newBlockTime - polo.Conf.BlockInterval
	for i := uint64(0); i < polo.Conf.MaxBlockProposers && t > s.parentBlockTime; i++ {
		p := s.whoseTurn(t)
		if p.Address != s.proposer.Address {
			toDeactivate[p.Address] = p
		}
		t -= polo.Conf.BlockInterval
	}

	updates = make([]Proposer, 0, len(toDeactivate)+1)
	for _, p := range toDeactivate {
		p.Active = false
		updates = append(updates, p)
	}

	if !s.proposer.Active {
		cpy := s.proposer
		cpy.Active = true
		updates = append(updates, cpy)
	}

	score = uint64(len(s.actives)) - uint64(len(toDeactivate))
	return
}

// dprp deterministic pseudo-random process.
// H(B, t)[:8]
func dprp(blockNumber uint32, time uint64) uint64 {
	var (
		b4 [4]byte
		b8 [8]byte
	)
	binary.BigEndian.PutUint32(b4[:], blockNumber)
	binary.BigEndian.PutUint64(b8[:], time)

	return binary.BigEndian.Uint64(polo.Blake2b(b4[:], b8[:]).Bytes())
}
