// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package state

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/storage/kv"
	"github.com/HiNounou029/nounouchain/trie"
)

// Stage abstracts changes on the main accounts trie.
type Stage struct {
	err error

	kv           kv.GetPutter
	accountTrie  *trie.SecureTrie
	storageTries []*trie.SecureTrie
	codes        []codeWithHash
}

type codeWithHash struct {
	code []byte
	hash []byte
}

func newStage(root polo.Bytes32, kv kv.GetPutter, changes map[polo.Address]*changedObject) *Stage {

	accountTrie, err := trCache.Get(root, kv, true)
	if err != nil {
		return &Stage{err: err}
	}

	storageTries := make([]*trie.SecureTrie, 0, len(changes))
	codes := make([]codeWithHash, 0, len(changes))

	for addr, obj := range changes {
		dataCpy := obj.data

		if len(obj.code) > 0 {
			codes = append(codes, codeWithHash{
				code: obj.code,
				hash: dataCpy.CodeHash})
		}

		// skip storage changes if account is empty
		if !dataCpy.IsEmpty() {
			if len(obj.storage) > 0 {
				strie, err := trCache.Get(polo.BytesToBytes32(dataCpy.StorageRoot), kv, true)
				if err != nil {
					return &Stage{err: err}
				}
				storageTries = append(storageTries, strie)
				for k, v := range obj.storage {
					if err := saveStorage(strie, k, v); err != nil {
						return &Stage{err: err}
					}
				}
				dataCpy.StorageRoot = strie.Hash().Bytes()
			}
		}

		if err := saveAccount(accountTrie, addr, &dataCpy); err != nil {
			return &Stage{err: err}
		}
	}
	return &Stage{
		kv:           kv,
		accountTrie:  accountTrie,
		storageTries: storageTries,
		codes:        codes,
	}
}

// Hash computes hash of the main accounts trie.
func (s *Stage) Hash() (polo.Bytes32, error) {
	if s.err != nil {
		return polo.Bytes32{}, s.err
	}
	return s.accountTrie.Hash(), nil
}

// Commit commits all changes into main accounts trie and storage tries.
func (s *Stage) Commit() (polo.Bytes32, error) {
	if s.err != nil {
		return polo.Bytes32{}, s.err
	}
	batch := s.kv.NewBatch()
	// write codes
	for _, code := range s.codes {
		if err := batch.Put(code.hash, code.code); err != nil {
			return polo.Bytes32{}, err
		}
	}

	// commit storage tries
	for _, strie := range s.storageTries {
		root, err := strie.CommitTo(batch)
		if err != nil {
			return polo.Bytes32{}, err
		}
		trCache.Add(root, strie, s.kv)
	}

	// commit accounts trie
	root, err := s.accountTrie.CommitTo(batch)
	if err != nil {
		return polo.Bytes32{}, err
	}

	if err := batch.Write(); err != nil {
		return polo.Bytes32{}, err
	}

	trCache.Add(root, s.accountTrie, s.kv)

	return root, nil
}
