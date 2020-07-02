// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package contract

import (
	"bytes"
	"encoding/json"
	"github.com/HiNounou029/nounouchain/poloclient"
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"io/ioutil"
	"math/big"
	"testing"
)

const (
	localEndpoint  = "http://localhost:8669"
	remoteEndpoint = "http://cefa_1:8669"
	abiFileName    = "bridge.json"
	localChainId   = 1000
	remoteChainId  = 2000
)

var localChainContractAddr = poloclient.MustParseAddress("0x233b177d891478c29df0e80d59e82839cba2c662")
var remoteChainContractAddr = poloclient.MustParseAddress("0xc5ea599b271e679ab2b8163663c30ad28dd4229f")

var keyFilePath = "/Users/majunchang/sys/cefa/bin/account1.key"

func TestABI(t *testing.T) {
	abiFileName := "bridge.json"
	data, err := ioutil.ReadFile(abiFileName)

	if err != nil {
		t.Fatalf("open file %s: %v", abiFileName, err)
	}

	ethABI, err := ethabi.JSON(bytes.NewReader(data))

	if err != nil {
		t.Fatalf("invalid abi content %s: %v", data, err)
	}

	t.Logf("methods %v", ethABI.Methods)

}

func newContract(t *testing.T, local bool) *BridgeContract {
	endPoint := localEndpoint
	if !local {
		endPoint = remoteEndpoint
	}
	bc, _ := poloclient.NewClient(endPoint, keyFilePath)
	abiBytes, err := ioutil.ReadFile(abiFileName)
	if err != nil {
		t.Fatalf("open file %s: %v", abiFileName, err)
	}

	contractAddr := localChainContractAddr
	if !local {
		contractAddr = remoteChainContractAddr
	}

	contract, err := NewBridgeContract(bc, abiBytes, contractAddr)
	if err != nil {
		t.Fatalf("new contract err %v", err)
	}
	return contract
}

func TestSend(t *testing.T) {
	contract := newContract(t, true)
	srcAddr := poloclient.MustParseAddress("0x2849B75972ae24f72BEC0f8d6f50AC0f3C8AfC21")
	dstChain := big.NewInt(remoteChainId)
	dstAddr := poloclient.MustParseAddress("0xaAd4b56142dA5c62B106ec1F691E2c28F479D0d8")
	reqMsg := "hello world"
	value := big.NewInt(0)

	txId, err := contract.Send(srcAddr, dstChain, dstAddr, value, reqMsg)
	if err != nil {
		t.Errorf("send error %v", err)
	}
	t.Logf("txId %s", txId)
}

func TestSendRemote(t *testing.T) {
	contract := newContract(t, false)
	srcAddr := poloclient.MustParseAddress("0x2849B75972ae24f72BEC0f8d6f50AC0f3C8AfC21")
	dstChain := big.NewInt(0)
	dstAddr := poloclient.MustParseAddress("0xaAd4b56142dA5c62B106ec1F691E2c28F479D0d8")
	reqMsg := "hello world"
	value := big.NewInt(0)

	txId, err := contract.Send(srcAddr, dstChain, dstAddr, value, reqMsg)
	if err != nil {
		t.Errorf("send error %v", err)
	}
	t.Logf("txId %s", txId)
}

func TestForward(t *testing.T) {
	localContract := newContract(t, true)
	msgId := big.NewInt(0)
	req, err := localContract.SendMsgsByDestChain(big.NewInt(remoteChainId), msgId)

	if err != nil {
		t.Fatal(err)
	}

	remoteContract := newContract(t, false)
	txId, err := remoteContract.Forward(big.NewInt(localChainId), big.NewInt(msgId.Int64()+1), req)

	if err != nil {
		t.Errorf("foward error %v", err)
	}
	t.Logf("txId %s", txId)
}

func TestAck(t *testing.T) {
	remoteContract := newContract(t, false)
	msgId := big.NewInt(0)
	res, err := remoteContract.RecvMsgsBySrcChain(big.NewInt(localChainId), msgId)

	if err != nil {
		t.Fatal(err)
	}

	localContract := newContract(t, true)

	txId, err := localContract.Ack(big.NewInt(remoteChainId),
		big.NewInt(msgId.Int64()+1), res.Success, res.ResMsg)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("txId %s", txId)
}

func TestLastMsgIdBySrcChain(t *testing.T) {
	remoteContract := newContract(t, false)
	msgId, err := remoteContract.LastMsgIdBySrcChain(big.NewInt(localChainId))

	if err != nil {
		t.Fatal(err)
	}
	t.Logf("msgId %d", msgId)
}

func TestRecvMsgsBySrcChain(t *testing.T) {
	remoteContract := newContract(t, false)
	msgId := big.NewInt(0)
	res, err := remoteContract.RecvMsgsBySrcChain(big.NewInt(localChainId), msgId)

	if err != nil {
		t.Fatal(err)
	}
	jsRes, err := json.MarshalIndent(res, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("response %s", jsRes)

}

func TestLastSentMsgIdByDestChain(t *testing.T) {
	contract := newContract(t, true)

	lastId, err := contract.LastSentMsgIdByDestChain(big.NewInt(remoteChainId))

	if err != nil {
		t.Fatal(err)
	}
	t.Logf("LastSentMsgId is %d", lastId)
}

func TestLastAckedMsgIdByDestChain(t *testing.T) {
	contract := newContract(t, true)

	lastId, err := contract.LastAckedMsgIdByDestChain(big.NewInt(remoteChainId))

	if err != nil {
		t.Fatal(err)
	}
	t.Logf("LastAckedMsgId is %d", lastId)
}

func TestSendMsgsByDestChain(t *testing.T) {
	contract := newContract(t, true)

	req, err := contract.SendMsgsByDestChain(big.NewInt(remoteChainId), big.NewInt(0))

	if err != nil {
		t.Fatal(err)
	}
	jsData, _ := json.MarshalIndent(req, "", "\t")
	t.Logf("%s", jsData)
}
