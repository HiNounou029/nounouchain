// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package accounts_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/HiNounou029/nounouchain/api/accounts"
	"github.com/HiNounou029/nounouchain/polo"
	ABI "github.com/HiNounou029/nounouchain/nounou/abi"
	"github.com/HiNounou029/nounouchain/nounou/genesis"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/HiNounou029/nounouchain/miner"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var sol = `	pragma solidity ^0.4.18;
			contract Test {
    			uint8 value;
    			function add(uint8 a,uint8 b) public pure returns(uint8) {
        			return a+b;
    			}
    			function set(uint8 v) public {
        			value = v;
    			}
			}`

var abiJSON = `[{
	"constant": true,
	"inputs": [{
		"name": "a",
		"type": "uint8"
	}, {
		"name": "b",
		"type": "uint8"
	}],
	"name": "add",
	"outputs": [{
		"name": "",
		"type": "uint8"
	}],
	"payable": false,
	"stateMutability": "pure",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"name": "v",
		"type": "uint8"
	}],
	"name": "set",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}]`
var addr = polo.BytesToAddress([]byte("to"))
var value = big.NewInt(10000)
var storageKey = polo.Bytes32{}
var storageValue = byte(1)

var contractAddr polo.Address

var bytecode = common.Hex2Bytes("608060405234801561001057600080fd5b50610125806100206000396000f3006080604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063378d92b414604e578063904e64181460a2575b600080fd5b348015605957600080fd5b506086600480360381019080803560ff169060200190929190803560ff16906020019092919050505060cf565b604051808260ff1660ff16815260200191505060405180910390f35b34801560ad57600080fd5b5060cd600480360381019080803560ff16906020019092919050505060dc565b005b6000818301905092915050565b806000806101000a81548160ff021916908360ff160217905550505600a165627a7a72305820a9c606a7fd6d33763c1392bdd81c4f496679f2daa736a27e5fff42881162287e0029")

var runtimeBytecode = common.Hex2Bytes("6080604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063378d92b414604e578063904e64181460a2575b600080fd5b348015605957600080fd5b506086600480360381019080803560ff169060200190929190803560ff16906020019092919050505060cf565b604051808260ff1660ff16815260200191505060405180910390f35b34801560ad57600080fd5b5060cd600480360381019080803560ff16906020019092919050505060dc565b005b6000818301905092915050565b806000806101000a81548160ff021916908360ff160217905550505600a165627a7a72305820a9c606a7fd6d33763c1392bdd81c4f496679f2daa736a27e5fff42881162287e0029")

var invalidAddr = "abc"                                                                   //invlaid address
var invalidBytes32 = "0x000000000000000000000000000000000000000000000000000000000000000g" //invlaid bytes32
var invalidNumberRevision = "4294967296"                                                  //invalid block number

var ts *httptest.Server

func TestAccount(t *testing.T) {
	initAccountServer(t)
	defer ts.Close()
	getAccount(t)
	getCode(t)
	getStorage(t)
	deployContractWithCall(t)
	callContract(t)
	batchCall(t)
}

func getAccount(t *testing.T) {
	res, statusCode := httpGet(t, ts.URL+"/accounts/"+invalidAddr)
	assert.Equal(t, http.StatusBadRequest, statusCode, "bad address")

	res, statusCode = httpGet(t, ts.URL+"/accounts/"+addr.String()+"?revision="+invalidNumberRevision)
	assert.Equal(t, http.StatusBadRequest, statusCode, "bad revision")

	//revision is optional defaut `best`
	res, statusCode = httpGet(t, ts.URL+"/accounts/"+addr.String())
	var acc accounts.Account
	if err := json.Unmarshal(res, &acc); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, math.HexOrDecimal256(*value), acc.Balance, "balance should be equal")
	assert.Equal(t, http.StatusOK, statusCode, "OK")

}

func getCode(t *testing.T) {
	res, statusCode := httpGet(t, ts.URL+"/accounts/"+invalidAddr+"/code")
	assert.Equal(t, http.StatusBadRequest, statusCode, "bad address")

	res, statusCode = httpGet(t, ts.URL+"/accounts/"+contractAddr.String()+"/code?revision="+invalidNumberRevision)
	assert.Equal(t, http.StatusBadRequest, statusCode, "bad revision")

	//revision is optional defaut `best`
	res, statusCode = httpGet(t, ts.URL+"/accounts/"+contractAddr.String()+"/code")
	var code map[string]string
	if err := json.Unmarshal(res, &code); err != nil {
		t.Fatal(err)
	}
	c, err := hexutil.Decode(code["code"])
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, runtimeBytecode, c, "code should be equal")
	assert.Equal(t, http.StatusOK, statusCode, "OK")
}

func getStorage(t *testing.T) {
	res, statusCode := httpGet(t, ts.URL+"/accounts/"+invalidAddr+"/storage/"+storageKey.String())
	assert.Equal(t, http.StatusBadRequest, statusCode, "bad address")

	res, statusCode = httpGet(t, ts.URL+"/accounts/"+contractAddr.String()+"/storage/"+invalidBytes32)
	assert.Equal(t, http.StatusBadRequest, statusCode, "bad storage key")

	res, statusCode = httpGet(t, ts.URL+"/accounts/"+contractAddr.String()+"/storage/"+storageKey.String()+"?revision="+invalidNumberRevision)
	assert.Equal(t, http.StatusBadRequest, statusCode, "bad revision")

	//revision is optional defaut `best`
	res, statusCode = httpGet(t, ts.URL+"/accounts/"+contractAddr.String()+"/storage/"+storageKey.String())
	var value map[string]string
	if err := json.Unmarshal(res, &value); err != nil {
		t.Fatal(err)
	}
	h, err := polo.ParseBytes32(value["value"])
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, polo.BytesToBytes32([]byte{storageValue}), h, "storage should be equal")
	assert.Equal(t, http.StatusOK, statusCode, "OK")
}

func initAccountServer(t *testing.T) {
	db, _ := storage.NewMem()
	stateC := state.NewCreator(db)
	gene := genesis.NewDevnet()

	b, _, err := gene.Build(stateC)
	if err != nil {
		t.Fatal(err)
	}
	chain, _ := chain.New(db, b)
	claTransfer := tx.NewClause(&addr).WithValue(value)
	claDeploy := tx.NewClause(nil).WithData(bytecode)
	transaction := buildTxWithClauses(t, chain.Tag(), claTransfer, claDeploy)
	contractAddr = polo.CreateContractAddress(transaction.ID(), 1, 0)
	packTx(chain, stateC, transaction, t)

	method := "set"
	abi, err := ABI.New([]byte(abiJSON))
	m, _ := abi.MethodByName(method)
	input, err := m.EncodeInput(uint8(storageValue))
	if err != nil {
		t.Fatal(err)
	}
	claCall := tx.NewClause(&contractAddr).WithData(input)
	transactionCall := buildTxWithClauses(t, chain.Tag(), claCall)
	packTx(chain, stateC, transactionCall, t)

	router := mux.NewRouter()
	accounts.New(chain, stateC, math.MaxUint64).Mount(router, "/accounts")
	ts = httptest.NewServer(router)
}

func buildTxWithClauses(t *testing.T, chaiTag byte, clauses ...*tx.Clause) *tx.Transaction {
	builder := new(tx.Builder).
		ChainTag(chaiTag).
		Expiration(10).
		Gas(1000000)
	for _, c := range clauses {
		builder.Clause(c)
	}

	transaction := builder.Build()
	sig, err := crypto.Sign(transaction.SigningHash().Bytes(), genesis.DevAccounts()[0].PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	return transaction.WithSignature(sig)
}

func packTx(chain *chain.Chain, stateC *state.Creator, transaction *tx.Transaction, t *testing.T) {
	b := chain.BestBlock()
	miner := miner.New(chain, stateC, genesis.DevAccounts()[0].Address, &genesis.DevAccounts()[0].Address)
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
	if _, err := chain.AddBlock(b, receipts); err != nil {
		t.Fatal(err)
	}
}

func deployContractWithCall(t *testing.T) {
	badBody := &accounts.CallData{
		Gas:  10000000,
		Data: "abc",
	}
	res, statusCode := httpPost(t, ts.URL+"/accounts", badBody)
	assert.Equal(t, http.StatusBadRequest, statusCode, "bad data")

	reqBody := &accounts.CallData{
		Gas:  10000000,
		Data: hexutil.Encode(bytecode),
	}

	res, statusCode = httpPost(t, ts.URL+"/accounts?revision="+invalidNumberRevision, reqBody)
	assert.Equal(t, http.StatusBadRequest, statusCode, "bad revision")

	//revision is optional defaut `best`
	res, statusCode = httpPost(t, ts.URL+"/accounts", reqBody)
	var output *accounts.CallResult
	if err := json.Unmarshal(res, &output); err != nil {
		t.Fatal(err)
	}
	assert.False(t, output.Reverted)

}

func callContract(t *testing.T) {
	res, statusCode := httpPost(t, ts.URL+"/accounts/"+invalidAddr, nil)
	assert.Equal(t, http.StatusBadRequest, statusCode, "invalid address")

	badBody := &accounts.CallData{
		Data: "input",
	}
	res, statusCode = httpPost(t, ts.URL+"/accounts/"+contractAddr.String(), badBody)
	assert.Equal(t, http.StatusBadRequest, statusCode, "invalid input data")

	a := uint8(1)
	b := uint8(2)
	method := "add"
	abi, err := ABI.New([]byte(abiJSON))
	m, _ := abi.MethodByName(method)
	input, err := m.EncodeInput(a, b)
	if err != nil {
		t.Fatal(err)
	}
	reqBody := &accounts.CallData{
		Data: hexutil.Encode(input),
	}
	res, statusCode = httpPost(t, ts.URL+"/accounts/"+contractAddr.String(), reqBody)
	var output *accounts.CallResult
	if err = json.Unmarshal(res, &output); err != nil {
		t.Fatal(err)
	}
	data, err := hexutil.Decode(output.Data)
	if err != nil {
		t.Fatal(err)
	}
	var ret uint8
	err = m.DecodeOutput(data, &ret)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, a+b, ret)
}

func batchCall(t *testing.T) {
	badBody := &accounts.BatchCallData{
		Clauses: accounts.Clauses{
			accounts.Clause{
				To:    &contractAddr,
				Data:  "data1",
				Value: nil,
			},
			accounts.Clause{
				To:    &contractAddr,
				Data:  "data2",
				Value: nil,
			}},
	}
	res, statusCode := httpPost(t, ts.URL+"/accounts/*", badBody)
	assert.Equal(t, http.StatusBadRequest, statusCode, "invalid data")

	a := uint8(1)
	b := uint8(2)
	method := "add"
	abi, err := ABI.New([]byte(abiJSON))
	m, _ := abi.MethodByName(method)
	input, err := m.EncodeInput(a, b)
	if err != nil {
		t.Fatal(err)
	}
	reqBody := &accounts.BatchCallData{
		Clauses: accounts.Clauses{
			accounts.Clause{
				To:    &contractAddr,
				Data:  hexutil.Encode(input),
				Value: nil,
			},
			accounts.Clause{
				To:    &contractAddr,
				Data:  hexutil.Encode(input),
				Value: nil,
			}},
	}

	res, statusCode = httpPost(t, ts.URL+"/accounts/*?revision="+invalidNumberRevision, badBody)
	assert.Equal(t, http.StatusBadRequest, statusCode, "invalid revision")

	res, statusCode = httpPost(t, ts.URL+"/accounts/*", reqBody)
	var results accounts.BatchCallResults
	if err = json.Unmarshal(res, &results); err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		data, err := hexutil.Decode(result.Data)
		if err != nil {
			t.Fatal(err)
		}
		var ret uint8
		err = m.DecodeOutput(data, &ret)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, a+b, ret, "should be equal")
	}
	assert.Equal(t, http.StatusOK, statusCode)
}

func httpPost(t *testing.T, url string, body interface{}) ([]byte, int) {
	data, err := json.Marshal(body)
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
	return r, res.StatusCode
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