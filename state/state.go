// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package state

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/HiNounou029/nounouchain/state/stm"
	"github.com/HiNounou029/nounouchain/storage/kv"
	"github.com/HiNounou029/nounouchain/trie"
	"github.com/ethereum/go-ethereum/rlp"
)

// State manages the main accounts trie.
type State struct {
	root     polo.Bytes32 // root of initial accounts trie
	kv       kv.GetPutter
	trie     trieReader                     // the accounts trie reader
	cache    map[polo.Address]*cachedObject // cache of accounts trie
	sm       *stm.StackedMap                // keeps revisions of accounts state
	err      error
	setError func(err error)
}

// to constrain ability of trie
type trieReader interface {
	TryGet(key []byte) ([]byte, error)
}

// to constrain ability of trie
type trieWriter interface {
	TryUpdate(key, value []byte) error
	TryDelete(key []byte) error
}

// New create an state object.
func New(root polo.Bytes32, kv kv.GetPutter) (*State, error) {
	trie, err := trCache.Get(root, kv, false)
	if err != nil {
		return nil, err
	}

	state := State{
		root:  root,
		kv:    kv,
		trie:  trie,
		cache: make(map[polo.Address]*cachedObject),
	}
	state.setError = func(err error) {
		if state.err == nil {
			state.err = err
		}
	}
	state.sm = stm.New(func(key interface{}) (value interface{}, exist bool) {
		return state.cacheGetter(key)
	})
	return &state, nil
}

// Spawn create a new state object shares current state's underlying db.
// Also errors will be reported to current state.
func (s *State) Spawn(root polo.Bytes32) *State {
	newState, err := New(root, s.kv)
	if err != nil {
		s.setError(err)
		newState, _ = New(polo.Bytes32{}, s.kv)
	}
	newState.setError = s.setError
	return newState
}

// implements stackedmap.MapGetter
func (s *State) cacheGetter(key interface{}) (value interface{}, exist bool) {
	switch k := key.(type) {
	case polo.Address: // get account
		return &s.getCachedObject(k).data, true
	case codeKey: // get code
		co := s.getCachedObject(polo.Address(k))
		code, err := co.GetCode()
		if err != nil {
			s.setError(err)
			return []byte(nil), true
		}
		return code, true
	case storageKey: // get storage
		v, err := s.getCachedObject(k.addr).GetStorage(k.key)
		if err != nil {
			s.setError(err)
			return rlp.RawValue(nil), true
		}
		return v, true
	}
	panic(fmt.Errorf("unexpected key type %+v", key))
}

// build changes via journal of stackedMap.
func (s *State) changes() map[polo.Address]*changedObject {
	changes := make(map[polo.Address]*changedObject)

	// get or create changedObject
	getOrNewObj := func(addr polo.Address) *changedObject {
		if obj, ok := changes[addr]; ok {
			return obj
		}
		obj := &changedObject{data: s.getCachedObject(addr).data}
		changes[addr] = obj
		return obj
	}

	// traverse journal to build changes
	s.sm.Journal(func(k, v interface{}) bool {
		switch key := k.(type) {
		case polo.Address:
			getOrNewObj(key).data = *(v.(*Account))
		case codeKey:
			getOrNewObj(polo.Address(key)).code = v.([]byte)
		case storageKey:
			o := getOrNewObj(key.addr)
			if o.storage == nil {
				o.storage = make(map[polo.Bytes32]rlp.RawValue)
			}
			o.storage[key.key] = v.(rlp.RawValue)
		}
		// abort if error occurred
		return s.err == nil
	})
	return changes
}

func (s *State) getCachedObject(addr polo.Address) *cachedObject {
	if co, ok := s.cache[addr]; ok {
		return co
	}
	a, err := loadAccount(s.trie, addr)
	if err != nil {
		s.setError(err)
		return newCachedObject(s.kv, emptyAccount())
	}
	co := newCachedObject(s.kv, a)
	s.cache[addr] = co
	return co
}

// the returned account should not be modified
func (s *State) getAccount(addr polo.Address) *Account {
	v, _ := s.sm.Get(addr)
	return v.(*Account)
}

func (s *State) getAccountCopy(addr polo.Address) Account {
	return *s.getAccount(addr)
}

func (s *State) updateAccount(addr polo.Address, acc *Account) {
	s.sm.Put(addr, acc)
}

// Err returns first occurred error.
func (s *State) Err() error {
	return s.err
}

// GetBalance returns balance for the given address.
func (s *State) GetBalance(addr polo.Address) *big.Int {
	return s.getAccount(addr).Balance
}

// SetBalance set balance for the given address.
func (s *State) SetBalance(addr polo.Address, balance *big.Int) {
	cpy := s.getAccountCopy(addr)
	cpy.Balance = balance
	s.updateAccount(addr, &cpy)
}

// Add add amount of balance to given address.
// false is returned if amount <0.
func (s *State) AddBalance(addr polo.Address, amount *big.Int) bool {
	if amount.Sign() == 0 {
		return true
	}

	if amount.Sign() < 0 {
		return false
	}

	s.SetBalance(addr, new(big.Int).Add(s.GetBalance(addr), amount))
	return true
}

// Sub sub amount of balance from given address.
// False is returned if no enough balance or ammount<=0
func (s *State) SubBalance(addr polo.Address, amount *big.Int) bool {
	if amount.Sign() == 0 {
		return true
	}
	if amount.Sign() < 0 {
		return false
	}
	bal := s.GetBalance(addr)
	if bal.Cmp(amount) < 0 {
		return false
	}
	s.SetBalance(addr, new(big.Int).Sub(bal, amount))
	return true
}

// GetMaster get master for the given address.
// Master can move energy, manage users...
func (s *State) GetMaster(addr polo.Address) polo.Address {
	return polo.BytesToAddress(s.getAccount(addr).Master)
}

// SetMaster set master for the given address.
func (s *State) SetMaster(addr polo.Address, master polo.Address) {
	cpy := s.getAccountCopy(addr)
	if master.IsZero() {
		cpy.Master = nil
	} else {
		cpy.Master = master[:]
	}
	s.updateAccount(addr, &cpy)
}

// GetStorage returns storage value for the given address and key.
func (s *State) GetStorage(addr polo.Address, key polo.Bytes32) polo.Bytes32 {
	raw := s.GetRawStorage(addr, key)
	if len(raw) == 0 {
		return polo.Bytes32{}
	}
	kind, content, _, err := rlp.Split(raw)
	if err != nil {
		s.setError(err)
		return polo.Bytes32{}
	}
	if kind == rlp.List {
		// special case for rlp list, it should be customized storage value
		// return hash of raw data
		return polo.Blake2b(raw)
	}
	return polo.BytesToBytes32(content)
}

// SetStorage set storage value for the given address and key.
func (s *State) SetStorage(addr polo.Address, key, value polo.Bytes32) {
	if value.IsZero() {
		s.SetRawStorage(addr, key, nil)
		return
	}
	v, _ := rlp.EncodeToBytes(bytes.TrimLeft(value[:], "\x00"))
	s.SetRawStorage(addr, key, v)
}

// GetRawStorage returns storage value in rlp raw for given address and key.
func (s *State) GetRawStorage(addr polo.Address, key polo.Bytes32) rlp.RawValue {
	data, _ := s.sm.Get(storageKey{addr, key})
	return data.(rlp.RawValue)
}

// SetRawStorage set storage value in rlp raw.
func (s *State) SetRawStorage(addr polo.Address, key polo.Bytes32, raw rlp.RawValue) {
	s.sm.Put(storageKey{addr, key}, raw)
}

// EncodeStorage set storage value encoded by given enc method.
// Error returned by end will be absorbed by State instance.
func (s *State) EncodeStorage(addr polo.Address, key polo.Bytes32, enc func() ([]byte, error)) {
	raw, err := enc()
	if err != nil {
		s.setError(err)
		return
	}
	s.SetRawStorage(addr, key, raw)
}

// DecodeStorage get and decode storage value.
// Error returned by dec will be absorbed by State instance.
func (s *State) DecodeStorage(addr polo.Address, key polo.Bytes32, dec func([]byte) error) {
	raw := s.GetRawStorage(addr, key)
	if err := dec(raw); err != nil {
		s.setError(err)
	}
}

// GetCode returns code for the given address.
func (s *State) GetCode(addr polo.Address) []byte {
	v, _ := s.sm.Get(codeKey(addr))
	return v.([]byte)
}

// GetCodeHash returns code hash for the given address.
func (s *State) GetCodeHash(addr polo.Address) polo.Bytes32 {
	return polo.BytesToBytes32(s.getAccount(addr).CodeHash)
}

// SetCode set code for the given address.
func (s *State) SetCode(addr polo.Address, code []byte) {
	var codeHash []byte
	if len(code) > 0 {
		s.sm.Put(codeKey(addr), code)
		codeHash = crypto.Keccak256(code)
	} else {
		s.sm.Put(codeKey(addr), []byte(nil))
	}
	cpy := s.getAccountCopy(addr)
	cpy.CodeHash = codeHash
	s.updateAccount(addr, &cpy)
}

// Exists returns whether an account exists at the given address.
// See Account.IsEmpty()
func (s *State) Exists(addr polo.Address) bool {
	return !s.getAccount(addr).IsEmpty()
}

// Delete delete an account at the given address.
// That's set balance, energy and code to zero value.
func (s *State) Delete(addr polo.Address) {
	s.sm.Put(codeKey(addr), []byte(nil))
	s.updateAccount(addr, emptyAccount())
}

// NewCheckpoint makes a checkpoint of current state.
// It returns revision of the checkpoint.
func (s *State) NewCheckpoint() int {
	return s.sm.Push()
}

// RevertTo revert to checkpoint specified by revision.
func (s *State) RevertTo(revision int) {
	s.sm.PopTo(revision)
}

// BuildStorageTrie build up storage trie for given address with cumulative changes.
func (s *State) BuildStorageTrie(addr polo.Address) (*trie.SecureTrie, error) {
	acc := s.getAccount(addr)

	root := polo.BytesToBytes32(acc.StorageRoot)

	// retrieve a copied trie
	trie, err := trCache.Get(root, s.kv, true)
	if err != nil {
		return nil, err
	}
	// traverse journal to filter out storage changes for addr
	s.sm.Journal(func(k, v interface{}) bool {
		switch key := k.(type) {
		case storageKey:
			if key.addr == addr {
				saveStorage(trie, key.key, v.(rlp.RawValue))
			}
		}
		// abort if error occurred
		return s.err == nil
	})
	if s.err != nil {
		return nil, s.err
	}
	return trie, nil
}

// Stage makes a stage object to compute hash of trie or commit all changes.
func (s *State) Stage() *Stage {
	if s.err != nil {
		return &Stage{err: s.err}
	}
	changes := s.changes()
	if s.err != nil {
		return &Stage{err: s.err}
	}
	return newStage(s.root, s.kv, changes)
}

type (
	storageKey struct {
		addr polo.Address
		key  polo.Bytes32
	}
	codeKey       polo.Address
	changedObject struct {
		data    Account
		storage map[polo.Bytes32]rlp.RawValue
		code    []byte
	}
)
