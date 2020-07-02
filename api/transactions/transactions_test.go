// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package transactions_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/HiNounou029/nounouchain/api/transactions"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/genesis"
	"github.com/HiNounou029/nounouchain/nounou/logdb"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/core/txpool"
	"github.com/HiNounou029/nounouchain/miner"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var c *chain.Chain
var ts *httptest.Server
var transaction *tx.Transaction

func TestTransaction(t *testing.T) {
	initTransactionServer(t)
	defer ts.Close()
	getTx(t)
	getTxReceipt(t)
	senTx(t)
}

func getTx(t *testing.T) {
	res := httpGet(t, ts.URL+"/transactions/"+transaction.ID().String())
	var rtx *transactions.PlainTransaction
	if err := json.Unmarshal(res, &rtx); err != nil {
		t.Fatal(err)
	}
	checkTx(t, transaction, rtx)

	res = httpGet(t, ts.URL+"/transactions/"+transaction.ID().String()+"?raw=true")
	var rawTx map[string]interface{}
	if err := json.Unmarshal(res, &rawTx); err != nil {
		t.Fatal(err)
	}
	rlpTx, err := rlp.EncodeToBytes(transaction)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, hexutil.Encode(rlpTx), rawTx["raw"], "should be equal raw")
}

func getTxReceipt(t *testing.T) {
	r := httpGet(t, ts.URL+"/transactions/"+transaction.ID().String()+"/receipt")
	var receipt *transactions.Receipt
	if err := json.Unmarshal(r, &receipt); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, uint64(receipt.GasUsed), transaction.Gas(), "gas should be equal")
}

func senTx(t *testing.T) {
	var blockRef = tx.NewBlockRef(0)
	var chainTag = c.Tag()
	var expiration = uint32(10)
	var gas = uint64(21000)

	tx := new(tx.Builder).
		BlockRef(blockRef).
		ChainTag(chainTag).
		Expiration(expiration).
		Gas(gas).
		Build()
	sig, err := crypto.Sign(tx.SigningHash().Bytes(), genesis.DevAccounts()[0].PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	tx = tx.WithSignature(sig)
	rlpTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	res := httpPost(t, ts.URL+"/transactions", transactions.RawTx{Raw: hexutil.Encode(rlpTx)})
	var txObj map[string]string
	if err = json.Unmarshal(res, &txObj); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, tx.ID().String(), txObj["id"], "should be the same transaction id")

	unsignedTx := transactions.UnSignedTx{
		ChainTag:   chainTag,
		BlockRef:   hexutil.Encode(blockRef[:]),
		Expiration: expiration,
		Gas:        gas,
	}
	res = httpPost(t, ts.URL+"/transactions", unsignedTx)
	if err = json.Unmarshal(res, &txObj); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, tx.SigningHash().String(), txObj["signingHash"], "should be the same transaction signingHash")

	signedTx := transactions.SignedTx{
		UnSignedTx: unsignedTx,
		Signature:  hexutil.Encode(sig),
	}
	res = httpPost(t, ts.URL+"/transactions", signedTx)
	if err = json.Unmarshal(res, &txObj); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, tx.ID().String(), txObj["id"], "should be the same transaction id")
}

func httpPost(t *testing.T, url string, obj interface{}) []byte {
	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	r, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func initTransactionServer(t *testing.T) {
	logDB, err := logdb.NewMem()
	if err != nil {
		t.Fatal(err)
	}
	from := polo.BytesToAddress([]byte("from"))
	to := polo.BytesToAddress([]byte("to"))
	value := big.NewInt(10)
	header := new(block.Builder).Build().Header()
	count := 100
	for i := 0; i < count; i++ {
		transLog := &tx.Transfer{
			Sender:    from,
			Recipient: to,
			Amount:    value,
		}
		header = new(block.Builder).ParentID(header.ID()).Build().Header()
		if err := logDB.Prepare(header).ForTransaction(polo.Bytes32{}, from).
			Insert(nil, tx.Transfers{transLog}).Commit(); err != nil {
			t.Fatal(err)
		}
	}
	db, _ := storage.NewMem()
	stateC := state.NewCreator(db)
	gene := genesis.NewDevnet()

	b, _, err := gene.Build(stateC)
	if err != nil {
		t.Fatal(err)
	}
	c, _ = chain.New(db, b)
	addr := polo.BytesToAddress([]byte("to"))
	cla := tx.NewClause(&addr).WithValue(big.NewInt(10000))
	transaction = new(tx.Builder).
		ChainTag(c.Tag()).
		Expiration(10).
		Gas(21000).
		Nonce(1).
		Clause(cla).
		BlockRef(tx.NewBlockRef(0)).
		Build()

	sig, err := crypto.Sign(transaction.SigningHash().Bytes(), genesis.DevAccounts()[0].PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	transaction = transaction.WithSignature(sig)
	miner := miner.New(c, stateC, genesis.DevAccounts()[0].Address, &genesis.DevAccounts()[0].Address)
	flow, err := miner.Schedule(b.Header(), uint64(time.Now().Unix()))
	err = flow.Adopt(transaction)
	if err != nil {
		t.Fatal(err)
	}
	b, stage, receipts, err := flow.Pack(genesis.DevAccounts()[0].PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stage.Commit(); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddBlock(b, receipts); err != nil {
		t.Fatal(err)
	}
	router := mux.NewRouter()
	transactions.New(c, txpool.New(c, stateC, txpool.Options{Limit: 10000, LimitPerAccount: 16, MaxLifetime: 10 * time.Minute})).Mount(router, "/transactions")
	ts = httptest.NewServer(router)

}

func checkTx(t *testing.T, expectedTx *tx.Transaction, actualTx *transactions.PlainTransaction) {
	origin, err := expectedTx.Signer()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, origin, actualTx.Origin)
	assert.Equal(t, expectedTx.ID(), actualTx.ID)
	assert.Equal(t, expectedTx.Gas(), actualTx.Gas)
	c := expectedTx.Clauses()[0]
	assert.Equal(t, hexutil.Encode(c.Data()), actualTx.Data)
	assert.Equal(t, *c.Value(), big.Int(actualTx.Value))
	assert.Equal(t, c.To(), actualTx.To)
}

func httpGet(t *testing.T, url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	r, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	return r
}
