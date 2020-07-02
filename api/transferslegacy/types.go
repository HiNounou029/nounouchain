// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package transferslegacy

import (
	"github.com/HiNounou029/nounouchain/api/transactions"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/logdb"
	"github.com/ethereum/go-ethereum/common/math"
)

type FilteredTransfer struct {
	Sender    polo.Address          `json:"sender"`
	Recipient polo.Address          `json:"recipient"`
	Amount    *math.HexOrDecimal256 `json:"amount"`
	Meta      transactions.LogMeta  `json:"meta"`
}

func convertTransfer(transfer *logdb.Transfer) *FilteredTransfer {
	v := math.HexOrDecimal256(*transfer.Amount)
	return &FilteredTransfer{
		Sender:    transfer.Sender,
		Recipient: transfer.Recipient,
		Amount:    &v,
		Meta: transactions.LogMeta{
			BlockID:        transfer.BlockID,
			BlockNumber:    transfer.BlockNumber,
			BlockTimestamp: transfer.BlockTime,
			TxID:           transfer.TxID,
			TxOrigin:       transfer.TxOrigin,
		},
	}
}

type AddressSet struct {
	TxOrigin  *polo.Address //who send transaction
	Sender    *polo.Address //who transferred tokens
	Recipient *polo.Address //who recieved tokens
}

type TransferFilter struct {
	TxID        *polo.Bytes32
	AddressSets []*AddressSet
	Range       *logdb.Range
	Options     *logdb.Options
	Order       logdb.Order //default asc
}

func convertTransferFilter(tf *TransferFilter) *logdb.TransferFilter {
	t := &logdb.TransferFilter{
		TxID:    tf.TxID,
		Range:   tf.Range,
		Options: tf.Options,
		Order:   tf.Order,
	}
	transferCriterias := make([]*logdb.TransferCriteria, len(tf.AddressSets))
	for i, addressSet := range tf.AddressSets {
		transferCriterias[i] = &logdb.TransferCriteria{
			TxOrigin:  addressSet.TxOrigin,
			Sender:    addressSet.Sender,
			Recipient: addressSet.Recipient,
		}
	}
	t.CriteriaSet = transferCriterias
	return t
}
