// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package state

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/storage/kv"
	"github.com/HiNounou029/nounouchain/trie"
	lru "github.com/hashicorp/golang-lru"
)

var trCache = newTrieCache()

type trieCache struct {
	cache *lru.Cache
}

type trieCacheEntry struct {
	trie *trie.SecureTrie
	kv   kv.GetPutter
}

func newTrieCache() *trieCache {
	cache, _ := lru.New(256)
	return &trieCache{cache: cache}
}

// to get a trie for writing, copy should be set to true
func (tc *trieCache) Get(root polo.Bytes32, kv kv.GetPutter, copy bool) (*trie.SecureTrie, error) {

	if v, ok := tc.cache.Get(root); ok {
		entry := v.(*trieCacheEntry)
		if entry.kv == kv {
			if copy {
				return entry.trie.Copy(), nil
			}
			return entry.trie, nil
		}
	}
	tr, err := trie.NewSecure(root, kv, 16)
	if err != nil {
		return nil, err
	}
	tc.cache.Add(root, &trieCacheEntry{tr, kv})
	if copy {
		return tr.Copy(), nil
	}
	return tr, nil
}

func (tc *trieCache) Add(root polo.Bytes32, trie *trie.SecureTrie, kv kv.GetPutter) {
	tc.cache.Add(root, &trieCacheEntry{trie.Copy(), kv})
}
