// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package contract

import (
	"bytes"
	"fmt"
	blcfabi "github.com/HiNounou029/nounouchain/nounou/abi"
	"github.com/HiNounou029/nounouchain/poloclient"
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type BridgeContract struct {
	ethABI       *ethabi.ABI
	blcfABI      *blcfabi.ABI
	blcfClient   *poloclient.PoloClient
	contractAddr poloclient.Address
}

func NewBridgeContract(blcfClient *poloclient.PoloClient,
	abiBytes []byte, contractAddr poloclient.Address) (*BridgeContract, error) {
	c := &BridgeContract{blcfClient: blcfClient, contractAddr: contractAddr}

	ethABI, err := ethabi.JSON(bytes.NewReader(abiBytes))
	if err != nil {
		return nil, fmt.Errorf("invalid eth abi %s: %v", abiBytes, err)
	}
	c.ethABI = &ethABI

	blcfABI, err := blcfabi.New(abiBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid polo abi %s: %v", abiBytes, err)
	}
	c.blcfABI = blcfABI

	return c, nil
}

func (c *BridgeContract) ClientAddr() *poloclient.Address {
	return c.blcfClient.ClientAddr()
}

func (c *BridgeContract) Send(srcAddr poloclient.Address, dstChain *big.Int,
	dstAddr poloclient.Address, value *big.Int, reqMsg string) ([]byte, error) {
	return c.blcfClient.InvokeContract(c.ethABI, c.contractAddr, value,
		"send", srcAddr, dstChain, dstAddr, reqMsg)
}

type Request struct {
	SrcAddr  common.Address
	DestAddr common.Address
	Value    *big.Int
	ReqMsg   string
	Status   int8
	ResMsg   string
}

type Response struct {
	SrcAddr  common.Address
	DestAddr common.Address
	Value    *big.Int
	Success  bool
	ResMsg   string
}

func (c *BridgeContract) Forward(srcChain *big.Int, msgId *big.Int, req *Request) ([]byte, error) {
	return c.blcfClient.InvokeContract(c.ethABI, c.contractAddr, req.Value,
		"forward", srcChain, msgId, req.SrcAddr, req.DestAddr, req.ReqMsg)
}

func (c *BridgeContract) Ack(destChain *big.Int, msgId *big.Int, success bool, resMsg string) ([]byte, error) {
	return c.blcfClient.InvokeContract(c.ethABI, c.contractAddr, big.NewInt(0),
		"ack", destChain, msgId, success, resMsg)
}

func (c *BridgeContract) LastSentMsgIdByDestChain(destChainId *big.Int) (*big.Int, error) {
	result := big.NewInt(0)
	err := c.blcfClient.CallContract(c.ethABI, c.contractAddr, &result,
		"lastSentMsgIdByDestChain", destChainId)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *BridgeContract) LastAckedMsgIdByDestChain(dstChainId *big.Int) (*big.Int, error) {
	result := big.NewInt(0)
	err := c.blcfClient.CallContract(c.ethABI, c.contractAddr, &result,
		"lastAckedMsgIdByDestChain", dstChainId)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *BridgeContract) LastMsgIdBySrcChain(srcChainId *big.Int) (*big.Int, error) {
	result := big.NewInt(0)
	err := c.blcfClient.CallContract(c.ethABI, c.contractAddr, &result,
		"lastMsgIdBySrcChain", srcChainId)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *BridgeContract) SendMsgsByDestChain(dstChainId *big.Int, msgId *big.Int) (*Request, error) {
	result := &Request{}
	err := c.blcfClient.CallContract(c.ethABI, c.contractAddr, result,
		"sendMsgsByDestChain", dstChainId, msgId)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *BridgeContract) RecvMsgsBySrcChain(srcChainId *big.Int, msgId *big.Int) (*Response, error) {
	result := &Response{}
	err := c.blcfClient.CallContract(c.ethABI, c.contractAddr, result,
		"recvMsgsBySrcChain", srcChainId, msgId)
	if err != nil {
		return nil, err
	}
	return result, nil
}
