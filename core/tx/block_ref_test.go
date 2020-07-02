// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package tx_test

import (
	"math/rand"
	"testing"

	"github.com/HiNounou029/nounouchain/polo"

	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/stretchr/testify/assert"
)

func TestBlockRef(t *testing.T) {
	assert.Equal(t, uint32(0), tx.BlockRef{}.Number())

	assert.Equal(t, tx.BlockRef{0, 0, 0, 0xff, 0, 0, 0, 0}, tx.NewBlockRef(0xff))

	var bid polo.Bytes32
	rand.Read(bid[:])

	br := tx.NewBlockRefFromID(bid)
	assert.Equal(t, bid[:8], br[:])
}
