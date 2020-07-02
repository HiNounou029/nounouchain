// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package polo

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/params"
)

// Constants of block chain.
const (
//	BlockInterval uint64 = 5 // time interval between two consecutive blocks.

	TxGas                     uint64 = 5000
	ClauseGas                 uint64 = params.TxGas - TxGas
	ClauseGasContractCreation uint64 = params.TxGasContractCreation - TxGas

	MinGasLimit          uint64 = 1000 * 1000
	InitialGasLimit      uint64 = 2 * 1000 * 1000 * 1000 // InitialGasLimit gas limit value int genesis block. 容纳大概10万笔交易
	GasLimitBoundDivisor uint64 = 1024                   // from ethereum
	GetBalanceGas        uint64 = 400                    //EIP158 gas table
	SloadGas             uint64 = 200                    // EIP158 gas table
	SstoreSetGas         uint64 = params.SstoreSetGas
	SstoreResetGas       uint64 = params.SstoreResetGas

	MaxTxWorkDelay uint32 = 30 // (unit: block) if tx delay exceeds this value, no energy can be exchanged.


	//TolerableBlockPackingTime = 100 * time.Millisecond // the indicator to adjust target block gas limit

	TolerableBlockPackingTime = 2 * time.Second // the indicator to adjust target block gas limit

	MaxBackTrackingBlockNumber = 65535
)

// Keys of governance params.
var (
	KeyExecutorAddress     = BytesToBytes32([]byte("executor"))
	KeyBaseGasPrice        = BytesToBytes32([]byte("base-gas-price"))
	KeyProposerEndorsement = BytesToBytes32([]byte("proposer-endorsement"))

	InitialBaseGasPrice        = big.NewInt(0)  // gas price设置为最小值

	InitialProposerEndorsement = new(big.Int).Mul(big.NewInt(1e18), big.NewInt(25000000))

	ApiProtocol = 0
	ConfigId [32]byte
)

type configuration struct {
	BlockInterval uint64
	TxPerSecondLimit uint64
	TxSizeLimit uint64
	MaxBlockProposers uint64
}

var Conf = configuration{5,2000, 65536, 7}

