// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package accounts

import (
	"github.com/HiNounou029/nounouchain/api/transactions"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/vm/runtime"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
)

//Account for marshal account
type Account struct {
	Balance math.HexOrDecimal256 `json:"balance"`
	HasCode bool                 `json:"hasCode"`
}

//CallData represents contract-call body
type CallData struct {
	Value    *math.HexOrDecimal256 `json:"value"`
	Data     string                `json:"data"`
	Gas      uint64                `json:"gas"`
	GasPrice *math.HexOrDecimal256 `json:"gasPrice"`
	Caller   *polo.Address         `json:"caller"`
}

type CallResult struct {
	Data      string                   `json:"data"`
	Events    []*transactions.Event    `json:"events"`
	Transfers []*transactions.Transfer `json:"transfers"`
	GasUsed   uint64                   `json:"gasUsed"`
	Reverted  bool                     `json:"reverted"`
	VMError   string                   `json:"vmError"`
}

func convertCallResultWithInputGas(vo *runtime.Output, inputGas uint64) *CallResult {
	gasUsed := inputGas - vo.LeftOverGas
	var (
		vmError  string
		reverted bool
	)

	if vo.VMErr != nil {
		reverted = true
		vmError = vo.VMErr.Error()
	}

	events := make([]*transactions.Event, len(vo.Events))
	transfers := make([]*transactions.Transfer, len(vo.Transfers))

	for j, txEvent := range vo.Events {
		event := &transactions.Event{
			Address: txEvent.Address,
			Data:    hexutil.Encode(txEvent.Data),
		}
		event.Topics = make([]polo.Bytes32, len(txEvent.Topics))
		for k, topic := range txEvent.Topics {
			event.Topics[k] = topic
		}
		events[j] = event
	}
	for j, txTransfer := range vo.Transfers {
		transfer := &transactions.Transfer{
			Sender:    txTransfer.Sender,
			Recipient: txTransfer.Recipient,
			Amount:    (*math.HexOrDecimal256)(txTransfer.Amount),
		}
		transfers[j] = transfer
	}

	return &CallResult{
		Data:      hexutil.Encode(vo.Data),
		Events:    events,
		Transfers: transfers,
		GasUsed:   gasUsed,
		Reverted:  reverted,
		VMError:   vmError,
	}
}

type Clause struct {
	To    *polo.Address         `json:"to"`
	Value *math.HexOrDecimal256 `json:"value"`
	Data  string                `json:"data"`
}

//Clauses array of clauses.
type Clauses []Clause

//BatchCallData executes a batch of codes
type BatchCallData struct {
	Clauses  Clauses               `json:"clauses"`
	Gas      uint64                `json:"gas"`
	GasPrice *math.HexOrDecimal256 `json:"gasPrice"`
	Caller   *polo.Address         `json:"caller"`
}

type BatchCallResults []*CallResult
