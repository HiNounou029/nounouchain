// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package blocks_test

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/HiNounou029/nounouchain/api/blocks"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/genesis"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/miner"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	testAddress = "56e81f171bcc55a6ff8345e692c0f86e5b48e01a"
	testPrivHex = "efa321f290811731036e5eccd373114e5186d9fe419081f5a607231279d5ef01"
)

var blk *block.Block
var ts *httptest.Server

var invalidBytes32 = "0x000000000000000000000000000000000000000000000000000000000000000g" //invlaid bytes32
var invalidNumberRevision = "4294967296"                                                  //invalid block number

func TestBlock(t *testing.T) {
	initBlockServer(t)
	defer ts.Close()
	//invalid block id
	res, statusCode := httpGet(t, ts.URL+"/blocks/"+invalidBytes32)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	//invalid block number
	res, statusCode = httpGet(t, ts.URL+"/blocks/"+invalidNumberRevision)
	assert.Equal(t, http.StatusBadRequest, statusCode)

	res, statusCode = httpGet(t, ts.URL+"/blocks/"+blk.Header().ID().String())
	rb := new(blocks.Block)
	if err := json.Unmarshal(res, &rb); err != nil {
		t.Fatal(err)
	}
	checkBlock(t, blk, rb)
	assert.Equal(t, http.StatusOK, statusCode)

	res, statusCode = httpGet(t, ts.URL+"/blocks/1")
	if err := json.Unmarshal(res, &rb); err != nil {
		t.Fatal(err)
	}
	checkBlock(t, blk, rb)
	assert.Equal(t, http.StatusOK, statusCode)

	res, statusCode = httpGet(t, ts.URL+"/blocks/best")
	if err := json.Unmarshal(res, &rb); err != nil {
		t.Fatal(err)
	}
	checkBlock(t, blk, rb)
	assert.Equal(t, http.StatusOK, statusCode)

}

func initBlockServer(t *testing.T) {
	db, _ := storage.NewMem()
	stateC := state.NewCreator(db)
	gene := genesis.NewDevnet()

	b, _, err := gene.Build(stateC)
	if err != nil {
		t.Fatal(err)
	}
	chain, _ := chain.New(db, b)
	addr := polo.BytesToAddress([]byte("to"))
	cla := tx.NewClause(&addr).WithValue(big.NewInt(10000))
	tx := new(tx.Builder).
		ChainTag(chain.Tag()).
		Expiration(10).
		Gas(21000).
		Nonce(1).
		Clause(cla).
		BlockRef(tx.NewBlockRef(0)).
		Build()

	sig, err := crypto.Sign(tx.SigningHash().Bytes(), genesis.DevAccounts()[0].PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	tx = tx.WithSignature(sig)
	miner := miner.New(chain, stateC, genesis.DevAccounts()[0].Address, &genesis.DevAccounts()[0].Address)
	flow, err := miner.Schedule(b.Header(), uint64(time.Now().Unix()))
	if err != nil {
		t.Fatal(err)
	}
	err = flow.Adopt(tx)
	if err != nil {
		t.Fatal(err)
	}
	block, stage, receipts, err := flow.Pack(genesis.DevAccounts()[0].PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stage.Commit(); err != nil {
		t.Fatal(err)
	}
	if _, err := chain.AddBlock(block, receipts); err != nil {
		t.Fatal(err)
	}
	router := mux.NewRouter()
	blocks.New(chain).Mount(router, "/blocks")
	ts = httptest.NewServer(router)
	blk = block
}

func checkBlock(t *testing.T, expBl *block.Block, actBl *blocks.Block) {
	header := expBl.Header()
	assert.Equal(t, header.Number(), actBl.Number, "Number should be equal")
	assert.Equal(t, header.ID(), actBl.ID, "Hash should be equal")
	assert.Equal(t, header.ParentID(), actBl.ParentID, "ParentID should be equal")
	assert.Equal(t, header.Timestamp(), actBl.Timestamp, "Timestamp should be equal")
	assert.Equal(t, header.TotalScore(), actBl.TotalScore, "TotalScore should be equal")
	assert.Equal(t, header.GasLimit(), actBl.GasLimit, "GasLimit should be equal")
	assert.Equal(t, header.GasUsed(), actBl.GasUsed, "GasUsed should be equal")
	assert.Equal(t, header.Beneficiary(), actBl.Beneficiary, "Beneficiary should be equal")
	assert.Equal(t, header.TxsRoot(), actBl.TxsRoot, "TxsRoot should be equal")
	assert.Equal(t, header.StateRoot(), actBl.StateRoot, "StateRoot should be equal")
	assert.Equal(t, header.ReceiptsRoot(), actBl.ReceiptsRoot, "ReceiptsRoot should be equal")
	for i, tx := range expBl.Transactions() {
		assert.Equal(t, tx.ID(), actBl.Transactions[i], "txid should be equal")
	}

}

func httpGet(t *testing.T, url string) ([]byte, int) {
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	r, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	return r, res.StatusCode
}
