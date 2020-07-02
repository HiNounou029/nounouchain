// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package transfers

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
	Reverted  bool                  `json:"reverted"`
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
		Reverted: transfer.Reverted,
	}
}
