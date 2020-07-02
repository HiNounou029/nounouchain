// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package poloclient

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/HiNounou029/nounouchain/api/status"
	"github.com/HiNounou029/nounouchain/api/transactions"
	"github.com/HiNounou029/nounouchain/api/utils"
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

type PoloClient struct {
	Endpoint    string
	Key         *ecdsa.PrivateKey
	ChainStatus *status.ChainStatus
	callerAddr  string
	clientAddr  *Address
}

func NewClient(endpoint string, keyFilePath string) (*PoloClient, error) {
	bc := &PoloClient{Endpoint: strings.TrimSuffix(endpoint, "/")}
	if len(keyFilePath) > 0 {
		key, err := crypto.LoadECDSA(keyFilePath)
		if err != nil {
			return nil, fmt.Errorf("open key file %s : %v", keyFilePath, err)
		}
		bc.Key = key
	}
	cs, err := bc.getChainStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get chain status: %v", err)
	}
	bc.ChainStatus = cs
	return bc, nil
}

func (bc* PoloClient) ClientAddr() *Address {
	if bc.Key == nil {
		return nil
	}
	if bc.clientAddr != nil {
		return bc.clientAddr
	}
	addr := Address(crypto.PubkeyToAddress(bc.Key.PublicKey))
	bc.clientAddr = &addr
	return bc.clientAddr
}

func (bc *PoloClient) caller() string {
	if bc.Key == nil {
		return ""
	}
	if bc.callerAddr != "" {
		return bc.callerAddr
	}
	addr := crypto.PubkeyToAddress(bc.Key.PublicKey)
	bc.callerAddr = addr.String()
	return bc.callerAddr
}

func (bc *PoloClient) getChainStatus() (*status.ChainStatus, error) {
	url := bc.Endpoint + "/status/"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("getting url %s: %v", url, err)
	}

	defer resp.Body.Close()
	status := &status.ChainStatus{}

	if err := utils.ParseJSON(resp.Body, &status); err != nil {
		return nil, fmt.Errorf("parse response of url %s: %v", url, err)
	}

	return status, nil
}

func (bc *PoloClient) GetChainStatus() *status.ChainStatus {
	return bc.ChainStatus
}

func (bc *PoloClient) GetStorageAt(addr string, key string) ([]byte, error) {
	url := bc.Endpoint + "/accounts/" + addr + "/storage/" + key
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("getting url %s: %v", url, err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read url body %s: %v", url, err)
	}
	return body, err
}

func (bc *PoloClient) GetBlock(index int) ([]byte, error) {
	url := bc.Endpoint + "/blocks/"
	if index == -1 {
		url += "best"
	} else {
		url += strconv.Itoa(index)
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get url %s: %v", url, err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read url body %s: %v", url, err)
	}
	return body, nil
}

func (bc *PoloClient) GetAccount(accountId string) ([]byte, error) {
	url := bc.Endpoint + "/accounts/" + accountId
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get url %s: %v", url, err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read url body %s: %v", url, err)
	}

	return body, nil
}

func (bc *PoloClient) GetTransaction(txId string) ([]byte, error) {
	url := bc.Endpoint + "/transactions/" + txId
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get url %s: %v", url, err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read url body %s: %v", url, err)
	}

	return body, nil
}

func (bc *PoloClient) GetTransactionReceipt(txId string) ([]byte, error) {
	url := bc.Endpoint + "/transactions/" + txId + "/receipt"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get url %s: %v", url, err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read url body %s: %v", url, err)
	}

	return body, nil
}

type JsonTx struct {
	To    string `json:"to"`
	Value uint64 `json:"value"`
	Data  string `json:"data"`
}

func httpPost(url string, obj interface{}) ([]byte, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("marshal %v: %v", obj, err)
	}
	res, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("post to %s with data %s: %v", url, data, err)
	}
	r, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read url %s: %v", url, err)
	}
	return r, nil
}

func (bc *PoloClient) SendTransaction(plain JsonTx) ([]byte, error) {
	addr := BytesToAddress(Hex2Bytes(plain.To))
	tx := NewPlainTransaction(bc, &addr, plain.Value, nil)

	if bc.Key == nil {
		return nil, errors.New("key is not initialized")
	}
	sig, err := crypto.Sign(tx.SigningHash().Bytes(), bc.Key)
	if err != nil {
		return nil, fmt.Errorf("sign tx error %v", err)
	}
	tx = tx.WithSignature(sig)
	rlpTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, fmt.Errorf("encode tx error %v", err)
	}
	res, err := httpPost(bc.Endpoint+"/transactions", PlainTx{Plain: hexutil.Encode(rlpTx)})
	if err != nil {
		return nil, fmt.Errorf("send tx %v: %v", tx, err)
	}
	return res, err
}

func (bc *PoloClient) DeployContract(data string) ([]byte, error) {
	tx := NewPlainTransaction(bc, nil, 0, Hex2Bytes(data))
	hash := tx.SigningHash().Bytes()
	if bc.Key == nil {
		return nil, errors.New("key is not initialized")
	}
	sig, err := crypto.Sign(hash, bc.Key)
	if err != nil {
		return nil, err
	}
	tx = tx.WithSignature(sig)
	rlpTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, fmt.Errorf("encode tx %v: %v", tx, err)
	}
	res, err := httpPost(bc.Endpoint+"/transactions", PlainTx{Plain: hexutil.Encode(rlpTx)})
	if err != nil {
		return nil, fmt.Errorf("send tx %v: %v", tx, err)
	}
	return res, err
}

//CallData represents contract-call body
type callData struct {
	Value    *math.HexOrDecimal256 `json:"value"`
	Data     string                `json:"data"`
	Gas      uint64                `json:"gas"`
	GasPrice *math.HexOrDecimal256 `json:"gasPrice"`
	Caller   string                `json:"caller"`
}

type callResult struct {
	Data      string                   `json:"data"`
	Events    []*transactions.Event    `json:"events"`
	Transfers []*transactions.Transfer `json:"transfers"`
	GasUsed   uint64                   `json:"gasUsed"`
	Reverted  bool                     `json:"reverted"`
	VMError   string                   `json:"vmError"`
}

func (bc *PoloClient) CallContract(
	ethABI *ethabi.ABI, contractAddr Address,
	v interface{},
	methodName string, args ...interface{}) error {
	data, err := ethABI.Pack(methodName, args...)
	if err != nil {
		return fmt.Errorf("pack method %s args %v: %v", methodName, args, err)
	}
	cdata := &callData{
		Value:    (*math.HexOrDecimal256)(big.NewInt(0)),
		Data:     hexutil.Encode(data),
		Gas:      uint64(1000000),
		GasPrice: (*math.HexOrDecimal256)(big.NewInt(0)),
		Caller:   bc.caller(),
	}
	result := callResult{}
	output, err := httpPost(bc.Endpoint+"/accounts/"+contractAddr.String(), cdata)
	if err != nil {
		return fmt.Errorf("call contract %v-%s: %v", contractAddr, methodName, err)
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return fmt.Errorf("invalid result %s: %v", output, err)
	}
	decodeData, err := hexutil.Decode(result.Data)
	if err != nil {
		return fmt.Errorf("decode result %s: %v", result.Data, err)
	}
	if err := ethABI.Unpack(v, methodName, decodeData); err != nil {
		return fmt.Errorf("unpack result %s,all result %s: %v", decodeData, output, err)
	}
	return nil
}

func (bc *PoloClient) InvokeContract(ethABI *ethabi.ABI,
	contractAddr Address, value *big.Int, methodName string, args ...interface{}) ([]byte, error) {
	txData, err := ethABI.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("pack methodName %s: %v", methodName, err)
	}
	tx := NewPlainTransaction(bc, &contractAddr, value.Uint64(), txData)
	if bc.Key == nil {
		return nil, errors.New("key is not initialized")
	}
	sig, err := crypto.Sign(tx.SigningHash().Bytes(), bc.Key)
	if err != nil {
		return nil, fmt.Errorf("sign tx error %v", err)
	}
	tx = tx.WithSignature(sig)
	rlpTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, fmt.Errorf("encode tx %v: %v", tx, err)
	}
	res, err := httpPost(bc.Endpoint+"/transactions", PlainTx{Plain: hexutil.Encode(rlpTx)})
	if err != nil {
		return nil, fmt.Errorf("send tx %v: %v", tx, err)
	}
	return res, err
}

func (bc *PoloClient) GetCode(addr string) ([]byte, error) {
	url := bc.Endpoint + "/accounts/" + addr + "/code"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get url %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read url %s: %v", url, err)
	}
	return body, nil
}
