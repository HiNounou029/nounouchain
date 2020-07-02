// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package block_test

import (
	"testing"
	"time"

	"github.com/HiNounou029/nounouchain/polo"
	. "github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)

func TestBlock(t *testing.T) {

	tx1 := new(tx.Builder).Clause(tx.NewClause(&polo.Address{})).Clause(tx.NewClause(&polo.Address{})).Build()
	tx2 := new(tx.Builder).Clause(tx.NewClause(nil)).Build()

	privKey := string("dce1443bd2ef0c2631adc1c67e5c93f13dc23a41c18b536effbbdcbcdb96fb65")

	now := uint64(time.Now().UnixNano())

	var (
		gasUsed     uint64       = 1000
		gasLimit    uint64       = 14000
		totalScore  uint64       = 101
		emptyRoot   polo.Bytes32 = polo.BytesToBytes32([]byte("0"))
		beneficiary polo.Address = polo.BytesToAddress([]byte("abc"))
	)

	block := new(Builder).
		GasUsed(gasUsed).
		Transaction(tx1).
		Transaction(tx2).
		GasLimit(gasLimit).
		TotalScore(totalScore).
		StateRoot(emptyRoot).
		ReceiptsRoot(emptyRoot).
		Timestamp(now).
		ParentID(emptyRoot).
		Beneficiary(beneficiary).
		Build()

	h := block.Header()

	txs := block.Transactions()
	body := block.Body()
	txsRootHash := txs.RootHash()

	//fmt.Println(h.ID())

	assert.Equal(t, body.Txs, txs)
	assert.Equal(t, Compose(h, txs), block)
	assert.Equal(t, gasLimit, h.GasLimit())
	assert.Equal(t, gasUsed, h.GasUsed())
	assert.Equal(t, totalScore, h.TotalScore())
	assert.Equal(t, emptyRoot, h.StateRoot())
	assert.Equal(t, emptyRoot, h.ReceiptsRoot())
	assert.Equal(t, now, h.Timestamp())
	assert.Equal(t, emptyRoot, h.ParentID())
	assert.Equal(t, beneficiary, h.Beneficiary())
	assert.Equal(t, txsRootHash, h.TxsRoot())

	key, _ := crypto.HexToECDSA(privKey)
	sig, _ := crypto.Sign(block.Header().SigningHash().Bytes(), key)

	block = block.WithSignature(sig)

	data, _ := rlp.EncodeToBytes(block)
	//fmt.Println(Raw(data).DecodeHeader())
	//fmt.Println(Raw(data).DecodeBody())

	//fmt.Println(block.Size())

	b := Block{}
	rlp.DecodeBytes(data, &b)
	//fmt.Println(b.Header().ID())
	//fmt.Println(&b)
}
