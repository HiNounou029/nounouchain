// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package subscriptions

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
)

//BlockMessage block piped by websocket
type BlockMessage struct {
	Number       uint32         `json:"number"`
	ID           polo.Bytes32   `json:"id"`
	Size         uint32         `json:"size"`
	ParentID     polo.Bytes32   `json:"parentID"`
	Timestamp    uint64         `json:"timestamp"`
	GasLimit     uint64         `json:"gasLimit"`
	Beneficiary  polo.Address   `json:"beneficiary"`
	GasUsed      uint64         `json:"gasUsed"`
	TotalScore   uint64         `json:"totalScore"`
	TxsRoot      polo.Bytes32   `json:"txsRoot"`
	StateRoot    polo.Bytes32   `json:"stateRoot"`
	ReceiptsRoot polo.Bytes32   `json:"receiptsRoot"`
	Signer       polo.Address   `json:"signer"`
	Transactions []polo.Bytes32 `json:"transactions"`
	Obsolete     bool           `json:"obsolete"`
}

func convertBlock(b *chain.Block) (*BlockMessage, error) {
	header := b.Header()
	signer, err := header.Signer()
	if err != nil {
		return nil, err
	}

	txs := b.Transactions()
	txIds := make([]polo.Bytes32, len(txs))
	for i, tx := range txs {
		txIds[i] = tx.ID()
	}
	return &BlockMessage{
		Number:       header.Number(),
		ID:           header.ID(),
		ParentID:     header.ParentID(),
		Timestamp:    header.Timestamp(),
		TotalScore:   header.TotalScore(),
		GasLimit:     header.GasLimit(),
		GasUsed:      header.GasUsed(),
		Beneficiary:  header.Beneficiary(),
		Signer:       signer,
		Size:         uint32(b.Size()),
		StateRoot:    header.StateRoot(),
		ReceiptsRoot: header.ReceiptsRoot(),
		TxsRoot:      header.TxsRoot(),
		Transactions: txIds,
		Obsolete:     b.Obsolete,
	}, nil
}

type LogMeta struct {
	BlockID        polo.Bytes32 `json:"blockID"`
	BlockNumber    uint32       `json:"blockNumber"`
	BlockTimestamp uint64       `json:"blockTimestamp"`
	TxID           polo.Bytes32 `json:"txID"`
	TxOrigin       polo.Address `json:"txOrigin"`
}

//TransferMessage transfer piped by websocket
type TransferMessage struct {
	Sender    polo.Address          `json:"sender"`
	Recipient polo.Address          `json:"recipient"`
	Amount    *math.HexOrDecimal256 `json:"amount"`
	Meta      LogMeta               `json:"meta"`
	Obsolete  bool                  `json:"obsolete"`
}

func convertTransfer(header *block.Header, tx *tx.Transaction, transfer *tx.Transfer, obsolete bool) (*TransferMessage, error) {
	signer, err := tx.Signer()
	if err != nil {
		return nil, err
	}

	return &TransferMessage{
		Sender:    transfer.Sender,
		Recipient: transfer.Recipient,
		Amount:    (*math.HexOrDecimal256)(transfer.Amount),
		Meta: LogMeta{
			BlockID:        header.ID(),
			BlockNumber:    header.Number(),
			BlockTimestamp: header.Timestamp(),
			TxID:           tx.ID(),
			TxOrigin:       signer,
		},
		Obsolete: obsolete,
	}, nil
}

//EventMessage event piped by websocket
type EventMessage struct {
	Address  polo.Address   `json:"address"`
	Topics   []polo.Bytes32 `json:"topics"`
	Data     string         `json:"data"`
	Meta     LogMeta        `json:"meta"`
	Obsolete bool           `json:"obsolete"`
}

func convertEvent(header *block.Header, tx *tx.Transaction, event *tx.Event, obsolete bool) (*EventMessage, error) {
	signer, err := tx.Signer()
	if err != nil {
		return nil, err
	}
	return &EventMessage{
		Address: event.Address,
		Data:    hexutil.Encode(event.Data),
		Meta: LogMeta{
			BlockID:        header.ID(),
			BlockNumber:    header.Number(),
			BlockTimestamp: header.Timestamp(),
			TxID:           tx.ID(),
			TxOrigin:       signer,
		},
		Topics:   event.Topics,
		Obsolete: obsolete,
	}, nil
}

// EventFilter contains options for contract event filtering.
type EventFilter struct {
	Address *polo.Address // restricts matches to events created by specific contracts
	Topic0  *polo.Bytes32
	Topic1  *polo.Bytes32
	Topic2  *polo.Bytes32
	Topic3  *polo.Bytes32
	Topic4  *polo.Bytes32
}

// Match returs whether event matches filter
func (ef *EventFilter) Match(event *tx.Event) bool {
	if (ef.Address != nil) && (*ef.Address != event.Address) {
		return false
	}

	matchTopic := func(topic *polo.Bytes32, index int) bool {
		if topic != nil {
			if len(event.Topics) <= index {
				return false
			}

			if *topic != event.Topics[index] {
				return false
			}
		}
		return true
	}

	return matchTopic(ef.Topic0, 0) &&
		matchTopic(ef.Topic1, 1) &&
		matchTopic(ef.Topic2, 2) &&
		matchTopic(ef.Topic3, 3) &&
		matchTopic(ef.Topic4, 4)
}

// TransferFilter contains options for contract transfer filtering.
type TransferFilter struct {
	TxOrigin  *polo.Address // who send transaction
	Sender    *polo.Address // who transferred tokens
	Recipient *polo.Address // who received tokens
}

// Match returs whether transfer matches filter
func (tf *TransferFilter) Match(transfer *tx.Transfer, origin polo.Address) bool {
	if (tf.TxOrigin != nil) && (*tf.TxOrigin != origin) {
		return false
	}

	if (tf.Sender != nil) && (*tf.Sender != transfer.Sender) {
		return false
	}

	if (tf.Recipient != nil) && (*tf.Recipient != transfer.Recipient) {
		return false
	}
	return true
}

type BeatMessage struct {
	Number    uint32       `json:"number"`
	ID        polo.Bytes32 `json:"id"`
	ParentID  polo.Bytes32 `json:"parentID"`
	Timestamp uint64       `json:"timestamp"`
	Bloom     string       `json:"bloom"`
	K         uint32       `json:"k"`
	Obsolete  bool         `json:"obsolete"`
}

type TranscationFilter struct {
	TxHash polo.Bytes32
}

func (tf *TranscationFilter) Match(tx *tx.Transaction) bool {

	if tf.TxHash != tx.ID() {
		return false
	}
	return true
}
