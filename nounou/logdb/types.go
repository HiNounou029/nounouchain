// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package logdb

import (
	"math/big"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/tx"
)

//Event represents tx.Event that can be stored in db.
type Event struct {
	BlockID     polo.Bytes32
	Index       uint32
	BlockNumber uint32
	BlockTime   uint64
	TxID        polo.Bytes32
	TxOrigin    polo.Address //contract caller
	Address     polo.Address // always a contract address
	Topics      [5]*polo.Bytes32
	Data        []byte
}

//newEvent converts tx.Event to Event.
func newEvent(header *block.Header, index uint32, txID polo.Bytes32, txOrigin polo.Address, txEvent *tx.Event) *Event {
	ev := &Event{
		BlockID:     header.ID(),
		Index:       index,
		BlockNumber: header.Number(),
		BlockTime:   header.Timestamp(),
		TxID:        txID,
		TxOrigin:    txOrigin,
		Address:     txEvent.Address, // always a contract address
		Data:        txEvent.Data,
	}
	for i := 0; i < len(txEvent.Topics) && i < len(ev.Topics); i++ {
		ev.Topics[i] = &txEvent.Topics[i]
	}
	return ev
}

//Transfer represents tx.Transfer that can be stored in db.
type Transfer struct {
	BlockID     polo.Bytes32
	Index       uint32
	BlockNumber uint32
	BlockTime   uint64
	TxID        polo.Bytes32
	TxOrigin    polo.Address
	Sender      polo.Address
	Recipient   polo.Address
	Amount      *big.Int
	Reverted    bool
}

//newTransfer converts tx.Transfer to Transfer.
func newTransfer(header *block.Header, index uint32, txID polo.Bytes32, txOrigin polo.Address, transfer *tx.Transfer, reverted bool) *Transfer {
	return &Transfer{
		BlockID:     header.ID(),
		Index:       index,
		BlockNumber: header.Number(),
		BlockTime:   header.Timestamp(),
		TxID:        txID,
		TxOrigin:    txOrigin,
		Sender:      transfer.Sender,
		Recipient:   transfer.Recipient,
		Amount:      transfer.Amount,
		Reverted:    reverted,
	}
}

type RangeType string

const (
	Block RangeType = "block"
	Time  RangeType = "time"
)

type Order string

const (
	ASC  Order = "asc"
	DESC Order = "desc"
)

type Range struct {
	Unit RangeType
	From uint64
	To   uint64
}

type Options struct {
	Offset uint64
	Limit  uint64
}

type EventCriteria struct {
	Address *polo.Address // always a contract address
	Topics  [5]*polo.Bytes32
}

//EventFilter filter
type EventFilter struct {
	CriteriaSet []*EventCriteria
	Range       *Range
	Options     *Options
	Order       Order //default asc
}

type TransferCriteria struct {
	TxOrigin  *polo.Address //who send transaction
	Sender    *polo.Address //who transferred tokens
	Recipient *polo.Address //who recieved tokens
}

type TransferFilter struct {
	TxID        *polo.Bytes32
	CriteriaSet []*TransferCriteria
	Range       *Range
	Options     *Options
	Order       Order //default asc
}
