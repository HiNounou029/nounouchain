// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package chain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockReader(t *testing.T) {
	ch := initChain()
	b0 := ch.GenesisBlock()

	b1 := newBlock(b0, 2)
	ch.AddBlock(b1, nil)

	b2 := newBlock(b1, 2)
	ch.AddBlock(b2, nil)

	b3 := newBlock(b2, 2)
	ch.AddBlock(b3, nil)

	b4 := newBlock(b3, 2)
	ch.AddBlock(b4, nil)

	br := ch.NewBlockReader(b2.Header().ID())

	blks, err := br.Read()
	assert.Nil(t, err)
	assert.Equal(t, blks[0].Header().ID(), b3.Header().ID())
	assert.False(t, blks[0].Obsolete)

	blks, err = br.Read()
	assert.Nil(t, err)
	assert.Equal(t, blks[0].Header().ID(), b4.Header().ID())
	assert.False(t, blks[0].Obsolete)
}

func TestBlockReaderFork(t *testing.T) {
	ch := initChain()
	b0 := ch.GenesisBlock()

	b1 := newBlock(b0, 2)
	ch.AddBlock(b1, nil)

	b2 := newBlock(b1, 2)
	ch.AddBlock(b2, nil)

	b2x := newBlock(b1, 1)
	ch.AddBlock(b2x, nil)

	b3 := newBlock(b2, 2)
	ch.AddBlock(b3, nil)

	b4 := newBlock(b3, 2)
	ch.AddBlock(b4, nil)

	br := ch.NewBlockReader(b2x.Header().ID())

	blks, err := br.Read()
	assert.Nil(t, err)
	assert.Equal(t, blks[0].Header().ID(), b2x.Header().ID())
	assert.True(t, blks[0].Obsolete)
	assert.Equal(t, blks[1].Header().ID(), b2.Header().ID())
	assert.False(t, blks[1].Obsolete)

	blks, err = br.Read()
	assert.Nil(t, err)
	assert.Equal(t, blks[0].Header().ID(), b3.Header().ID())
	assert.False(t, blks[0].Obsolete)
}
