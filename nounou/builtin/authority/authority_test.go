// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package authority

import (
	"math/big"
	"testing"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/stretchr/testify/assert"
)

func M(a ...interface{}) []interface{} {
	return a
}

func TestAuthority(t *testing.T) {
	kv, _ := storage.NewMem()
	st, _ := state.New(polo.Bytes32{}, kv)

	p1 := polo.BytesToAddress([]byte("p1"))
	p2 := polo.BytesToAddress([]byte("p2"))
	p3 := polo.BytesToAddress([]byte("p3"))

	st.SetBalance(p1, big.NewInt(10))
	st.SetBalance(p2, big.NewInt(20))
	st.SetBalance(p3, big.NewInt(30))

	aut := New(polo.BytesToAddress([]byte("aut")), st)
	tests := []struct {
		ret      interface{}
		expected interface{}
	}{
		{aut.Add(p1, p1, polo.Bytes32{}), true},
		{M(aut.Get(p1)), []interface{}{true, p1, polo.Bytes32{}, true}},
		{aut.Add(p2, p2, polo.Bytes32{}), true},
		{aut.Add(p3, p3, polo.Bytes32{}), true},
		{M(aut.Candidates(big.NewInt(10), polo.Conf.MaxBlockProposers)), []interface{}{
			[]*Candidate{{p1, p1, polo.Bytes32{}, true}, {p2, p2, polo.Bytes32{}, true}, {p3, p3, polo.Bytes32{}, true}},
		}},
		{M(aut.Candidates(big.NewInt(20), polo.Conf.MaxBlockProposers)), []interface{}{
			[]*Candidate{{p2, p2, polo.Bytes32{}, true}, {p3, p3, polo.Bytes32{}, true}},
		}},
		{M(aut.Candidates(big.NewInt(30), polo.Conf.MaxBlockProposers)), []interface{}{
			[]*Candidate{{p3, p3, polo.Bytes32{}, true}},
		}},
		{M(aut.Candidates(big.NewInt(10), 2)), []interface{}{
			[]*Candidate{{p1, p1, polo.Bytes32{}, true}, {p2, p2, polo.Bytes32{}, true}},
		}},
		{M(aut.Get(p1)), []interface{}{true, p1, polo.Bytes32{}, true}},
		{aut.Update(p1, false), true},
		{M(aut.Get(p1)), []interface{}{true, p1, polo.Bytes32{}, false}},
		{aut.Update(p1, true), true},
		{M(aut.Get(p1)), []interface{}{true, p1, polo.Bytes32{}, true}},
		{aut.Revoke(p1), true},
		{M(aut.Get(p1)), []interface{}{false, p1, polo.Bytes32{}, false}},
		{M(aut.Candidates(&big.Int{}, polo.Conf.MaxBlockProposers)), []interface{}{
			[]*Candidate{{p2, p2, polo.Bytes32{}, true}, {p3, p3, polo.Bytes32{}, true}},
		}},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.ret)
	}

	assert.Nil(t, st.Err())

}
