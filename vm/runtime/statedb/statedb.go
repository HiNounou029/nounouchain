// Copyright (c) 2019 The PoloChain Chain developers

// Distributed under the GNU Lesser General Public License v3.0 software license, see the accompanying
// file LICENSE or <https://www.gnu.org/licenses/lgpl-3.0.html>

package statedb

import (
	"fmt"
	"math/big"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/state/stm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	lru "github.com/hashicorp/golang-lru"
)

var codeSizeCache, _ = lru.New(32 * 1024)

// StateDB implements evm.StateDB, only adapt to evm.
type StateDB struct {
	state *state.State
	repo  *stm.StackedMap
}

type (
	suicideFlagKey common.Address
	refundKey      struct{}
	preimageKey    common.Hash
	eventKey       struct{}
	transferKey    struct{}
	stateRevKey    struct{}
)

// New create a statedb object.
func New(state *state.State) *StateDB {
	getter := func(k interface{}) (interface{}, bool) {
		switch k.(type) {
		case suicideFlagKey:
			return false, true
		case refundKey:
			return uint64(0), true
		}
		panic(fmt.Sprintf("unknown type of key %+v", k))
	}

	repo := stm.New(getter)
	return &StateDB{
		state,
		repo,
	}
}

// GetRefund returns total refund during VM life-cycle.
func (s *StateDB) GetRefund() uint64 {
	v, _ := s.repo.Get(refundKey{})
	return v.(uint64)
}

// GetLogs returns collected event and transfer logs.
func (s *StateDB) GetLogs() (tx.Events, tx.Transfers) {
	var (
		events    tx.Events
		transfers tx.Transfers
	)
	s.repo.Journal(func(k, v interface{}) bool {
		switch k.(type) {
		case eventKey:
			events = append(events, ethlogToEvent(v.(*types.Log)))
		case transferKey:
			transfers = append(transfers, v.(*tx.Transfer))
		}
		return true
	})
	return events, transfers
}

// ForEachStorage see state.State.ForEachStorage.
// func (s *StateDB) ForEachStorage(addr common.Address, cb func(common.Hash, common.Hash) bool) {
// 	s.state.ForEachStorage(polo.Address(addr), func(k polo.Bytes32, v []byte) bool {
// 		// TODO should rlp decode v
// 		return cb(common.Hash(k), common.BytesToHash(v))
// 	})
// }

// CreateAccount stub.
func (s *StateDB) CreateAccount(addr common.Address) {}

// GetBalance stub.
func (s *StateDB) GetBalance(addr common.Address) *big.Int {
	return s.state.GetBalance(polo.Address(addr))
}

// SubBalance stub.
func (s *StateDB) SubBalance(addr common.Address, amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	balance := s.state.GetBalance(polo.Address(addr))
	s.state.SetBalance(polo.Address(addr), new(big.Int).Sub(balance, amount))
}

// AddBalance stub.
func (s *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	balance := s.state.GetBalance(polo.Address(addr))
	s.state.SetBalance(polo.Address(addr), new(big.Int).Add(balance, amount))
}

// GetNonce stub.
func (s *StateDB) GetNonce(addr common.Address) uint64 { return 0 }

// SetNonce stub.
func (s *StateDB) SetNonce(addr common.Address, nonce uint64) {}

// GetCodeHash stub.
func (s *StateDB) GetCodeHash(addr common.Address) common.Hash {
	return common.Hash(s.state.GetCodeHash(polo.Address(addr)))
}

// GetCode stub.
func (s *StateDB) GetCode(addr common.Address) []byte {
	return s.state.GetCode(polo.Address(addr))
}

// GetCodeSize stub.
func (s *StateDB) GetCodeSize(addr common.Address) int {
	hash := s.state.GetCodeHash(polo.Address(addr))
	if hash.IsZero() {
		return 0
	}
	if v, ok := codeSizeCache.Get(hash); ok {
		return v.(int)
	}
	size := len(s.state.GetCode(polo.Address(addr)))
	codeSizeCache.Add(hash, size)
	return size
}

// SetCode stub.
func (s *StateDB) SetCode(addr common.Address, code []byte) {
	s.state.SetCode(polo.Address(addr), code)
}

// HasSuicided stub.
func (s *StateDB) HasSuicided(addr common.Address) bool {
	// only check suicide flag here
	v, _ := s.repo.Get(suicideFlagKey(addr))
	return v.(bool)
}

// Suicide stub.
// We do two things:
// 1, delete account
// 2, set suicide flag
func (s *StateDB) Suicide(addr common.Address) bool {
	if !s.state.Exists(polo.Address(addr)) {
		return false
	}
	s.state.Delete(polo.Address(addr))
	s.repo.Put(suicideFlagKey(addr), true)
	return true
}

// GetState stub.
func (s *StateDB) GetState(addr common.Address, key common.Hash) common.Hash {
	return common.Hash(s.state.GetStorage(polo.Address(addr), polo.Bytes32(key)))
}

// SetState stub.
func (s *StateDB) SetState(addr common.Address, key, value common.Hash) {
	s.state.SetStorage(polo.Address(addr), polo.Bytes32(key), polo.Bytes32(value))
}

// Exist stub.
func (s *StateDB) Exist(addr common.Address) bool {
	return s.state.Exists(polo.Address(addr))
}

// Empty stub.
func (s *StateDB) Empty(addr common.Address) bool {
	return !s.state.Exists(polo.Address(addr))
}

// AddRefund stub.
func (s *StateDB) AddRefund(gas uint64) {
	v, _ := s.repo.Get(refundKey{})
	total := v.(uint64) + gas
	s.repo.Put(refundKey{}, total)
}

// AddPreimage stub.
func (s *StateDB) AddPreimage(hash common.Hash, preimage []byte) {
	s.repo.Put(preimageKey(hash), preimage)
}

// AddLog stub.
func (s *StateDB) AddLog(vmlog *types.Log) {
	s.repo.Put(eventKey{}, vmlog)
}

func (s *StateDB) AddTransfer(transfer *tx.Transfer) {
	s.repo.Put(transferKey{}, transfer)
}

// Snapshot stub.
func (s *StateDB) Snapshot() int {
	srev := s.state.NewCheckpoint()
	s.repo.Put(stateRevKey{}, srev)
	return s.repo.Push()
}

// RevertToSnapshot stub.
func (s *StateDB) RevertToSnapshot(rev int) {
	s.repo.PopTo(rev)
	if srev, ok := s.repo.Get(stateRevKey{}); ok {
		s.state.RevertTo(srev.(int))
	} else {
		panic("state checkpoint missing")
	}
}

func ethlogToEvent(ethlog *types.Log) *tx.Event {
	var topics []polo.Bytes32
	if len(ethlog.Topics) > 0 {
		topics = make([]polo.Bytes32, 0, len(ethlog.Topics))
		for _, t := range ethlog.Topics {
			topics = append(topics, polo.Bytes32(t))
		}
	}
	return &tx.Event{
		polo.Address(ethlog.Address),
		topics,
		ethlog.Data,
	}
}
