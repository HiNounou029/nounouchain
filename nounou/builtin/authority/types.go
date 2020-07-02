// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package authority

import (
	"github.com/HiNounou029/nounouchain/polo"
)

type (
	entry struct {
		Endorsor polo.Address
		Identity polo.Bytes32
		Active   bool
		Prev     *polo.Address `rlp:"nil"`
		Next     *polo.Address `rlp:"nil"`
	}

	// Candidate candidate of block proposer.
	Candidate struct {
		NodeMaster polo.Address
		Endorsor   polo.Address
		Identity   polo.Bytes32
		Active     bool
	}
)

// IsEmpty returns whether the entry can be treated as empty.
func (e *entry) IsEmpty() bool {
	return e.Endorsor.IsZero() &&
		e.Identity.IsZero() &&
		!e.Active &&
		e.Prev == nil &&
		e.Next == nil
}

func (e *entry) IsLinked() bool {
	return e.Prev != nil || e.Next != nil
}
