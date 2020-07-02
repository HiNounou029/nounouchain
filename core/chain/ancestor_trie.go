// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package chain

import (
	"encoding/binary"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/storage/kv"
	"github.com/HiNounou029/nounouchain/trie"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

const rootCacheLimit = 2048

type ancestorTrie struct {
	kv         kv.GetPutter
	rootsCache *cache
	trieCache  *trieCache
}

func newAncestorTrie(kv kv.GetPutter) *ancestorTrie {
	rootsCache := newCache(rootCacheLimit, func(key interface{}) (interface{}, error) {
		return loadBlockNumberIndexTrieRoot(kv, key.(polo.Bytes32))
	})
	return &ancestorTrie{kv, rootsCache, newTrieCache()}
}

func numberAsKey(num uint32) []byte {
	var key [4]byte
	binary.BigEndian.PutUint32(key[:], num)
	return key[:]
}

func (at *ancestorTrie) Update(w kv.Putter, id, parentID polo.Bytes32) error {
	var parentRoot polo.Bytes32
	if block.Number(id) > 0 {
		// non-genesis
		root, err := at.rootsCache.GetOrLoad(parentID)
		if err != nil {
			return errors.WithMessage(err, "load index root")
		}
		parentRoot = root.(polo.Bytes32)
	}

	tr, err := at.trieCache.Get(parentRoot, at.kv, true)
	if err != nil {
		return err
	}

	if err := tr.TryUpdate(numberAsKey(block.Number(id)), id[:]); err != nil {
		return err
	}

	root, err := tr.CommitTo(w)
	if err != nil {
		return err
	}
	if err := saveBlockNumberIndexTrieRoot(w, id, root); err != nil {
		return err
	}
	at.trieCache.Add(root, tr, at.kv)
	at.rootsCache.Add(id, root)
	return nil
}

func (at *ancestorTrie) GetAncestor(descendantID polo.Bytes32, ancestorNum uint32) (polo.Bytes32, error) {
	if ancestorNum > block.Number(descendantID) {
		return polo.Bytes32{}, errNotFound
	}
	if ancestorNum == block.Number(descendantID) {
		return descendantID, nil
	}

	root, err := at.rootsCache.GetOrLoad(descendantID)
	if err != nil {
		return polo.Bytes32{}, errors.WithMessage(err, "load index root")
	}
	tr, err := at.trieCache.Get(root.(polo.Bytes32), at.kv, false)
	if err != nil {
		return polo.Bytes32{}, err
	}

	id, err := tr.TryGet(numberAsKey(ancestorNum))
	if err != nil {
		return polo.Bytes32{}, err
	}
	return polo.BytesToBytes32(id), nil
}

///
type trieCache struct {
	cache *lru.Cache
}

type trieCacheEntry struct {
	trie *trie.Trie
	kv   kv.GetPutter
}

func newTrieCache() *trieCache {
	cache, _ := lru.New(16)
	return &trieCache{cache: cache}
}

// to get a trie for writing, copy should be set to true
func (tc *trieCache) Get(root polo.Bytes32, kv kv.GetPutter, copy bool) (*trie.Trie, error) {

	if v, ok := tc.cache.Get(root); ok {
		entry := v.(*trieCacheEntry)
		if entry.kv == kv {
			if copy {
				cpy := *entry.trie
				return &cpy, nil
			}
			return entry.trie, nil
		}
	}
	tr, err := trie.New(root, kv)
	if err != nil {
		return nil, err
	}
	tr.SetCacheLimit(16)
	tc.cache.Add(root, &trieCacheEntry{tr, kv})
	if copy {
		cpy := *tr
		return &cpy, nil
	}
	return tr, nil
}

func (tc *trieCache) Add(root polo.Bytes32, trie *trie.Trie, kv kv.GetPutter) {
	cpy := *trie
	tc.cache.Add(root, &trieCacheEntry{&cpy, kv})
}
