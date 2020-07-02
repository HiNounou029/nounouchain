// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package poloclient_test

import (
	"bytes"
	"encoding/json"
	"github.com/HiNounou029/nounouchain/poloclient"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"io/ioutil"
	"math/big"
	"testing"
)

const (
	ENDPOINT    = "http://localhost:8669"
	keyFilePath = "/Users/majunchang/sys/cefa/bin/account1.key"
	//keyFilePath = "/Users/majunchang/sys/cefa/bin/poor.key"
)

func TestGetGenesis(t *testing.T) {
	bc, _ := poloclient.NewClient(ENDPOINT, keyFilePath)

	data, err := bc.GetBlock(0)

	if err != nil {
		t.Error(err)
	}
	t.Logf("%s", data)
}

func TestSendTx(t *testing.T) {
	bc, _ := poloclient.NewClient(ENDPOINT, keyFilePath)

	tx := poloclient.JsonTx{
		To:    "aad4b56142da5c62b106ec1f691e2c28f479d0d8",
		Value: 10000,
		Data:  "",
	}
	txId, err := bc.SendTransaction(tx)

	if err != nil {
		t.Error(err)
	}
	t.Logf("%s", string(txId))
}

func TestLastSentMsgIdByDestChain(t *testing.T) {
	bc, _ := poloclient.NewClient(ENDPOINT, keyFilePath)
	addr := poloclient.MustParseAddress("0x233b177d891478c29df0e80d59e82839cba2c662")
	abiFileName := "../cmd/bridge/contract/bridge.json"
	abiBytes, _ := ioutil.ReadFile(abiFileName)
	bridgeABI, err := abi.JSON(bytes.NewReader(abiBytes))
	_, err = bridgeABI.Pack("lastSentMsgIdByDestChain", big.NewInt(2000))
	if err != nil {
		t.Fatal(err)
	}
	result := big.NewInt(0)

	err = bc.CallContract(&bridgeABI, addr, &result,
		"lastSentMsgIdByDestChain", big.NewInt(2000))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}

func TestSendMsgIdByDestChain(t *testing.T) {
	bc, _ := poloclient.NewClient(ENDPOINT, keyFilePath)
	addr := poloclient.MustParseAddress("0x233b177d891478c29df0e80d59e82839cba2c662")
	abiFileName := "../cmd/bridge/contract/bridge.json"
	abiBytes, _ := ioutil.ReadFile(abiFileName)
	bridgeABI, err := abi.JSON(bytes.NewReader(abiBytes))
	result := struct {
		SrcAddr  common.Address
		DestAddr common.Address
		Value    *big.Int
		ReqMsg   string
		Status   int8
		ResMsg   string
	}{}
	err = bc.CallContract(&bridgeABI, addr, &result,
		"sendMsgsByDestChain", big.NewInt(2000), big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(&result, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", data)
}
