// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package chain

import (
	"testing"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)

func TestRawBlock(t *testing.T) {
	b := new(block.Builder).ParentID(polo.Bytes32{1, 2, 3}).Build()

	priv, _ := crypto.GenerateKey()
	sig, err := crypto.Sign(b.Header().SigningHash().Bytes(), priv)
	assert.Nil(t, err)
	b = b.WithSignature(sig)

	data, _ := rlp.EncodeToBytes(b)
	raw := &rawBlock{raw: data}

	h, _ := raw.Header()
	assert.Equal(t, b.Header().ID(), h.ID())

	b1, _ := raw.Block()

	data, _ = rlp.EncodeToBytes(b1)
	assert.Equal(t, []byte(raw.raw), data)
}
