// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package transactions

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
)

// Clause for json marshal
type Clause struct {
	To    *polo.Address        `json:"to"`
	Value math.HexOrDecimal256 `json:"value"`
	Data  string               `json:"data"`
}

//Clauses array of clauses.
type Clauses []Clause

//ConvertClause convert a raw clause into a json format clause
func convertClause(c *tx.Clause) Clause {
	return Clause{
		c.To(),
		math.HexOrDecimal256(*c.Value()),
		hexutil.Encode(c.Data()),
	}
}

func (c *Clause) String() string {
	return fmt.Sprintf(`Clause(
		To    %v
		Value %v
		Data  %v
		)`, c.To,
		c.Value,
		c.Data)
}

func hasKey(m map[string]interface{}, key string) bool {
	for k := range m {
		if strings.ToLower(k) == strings.ToLower(key) {
			return true
		}
	}
	return false
}

//Transaction transaction
type Transaction struct {
	ID           polo.Bytes32        `json:"id"`
	ChainTag     byte                `json:"chainTag"`
	BlockRef     string              `json:"blockRef"`
	Expiration   uint32              `json:"expiration"`
	Clauses      Clauses             `json:"clauses"`
	Gas          uint64              `json:"gas"`
	Origin       polo.Address        `json:"origin"`
	Nonce        math.HexOrDecimal64 `json:"nonce"`
	DependsOn    *polo.Bytes32       `json:"dependsOn"`
	Size         uint32              `json:"size"`
	Meta         TxMeta              `json:"meta"`
}

//Transaction transaction
type PlainTransaction struct {
	ID         polo.Bytes32         `json:"id"`
	ChainTag   byte                 `json:"chainTag"`
	Expiration uint32               `json:"expiration"`
	Gas        uint64               `json:"gas"`
	Origin     polo.Address         `json:"origin"`
	Nonce      math.HexOrDecimal64  `json:"nonce"`
	Meta       TxMeta               `json:"meta"`
	To         *polo.Address        `json:"to"`
	Value      math.HexOrDecimal256 `json:"value"`
	Data       string               `json:"data"`
}

type UnSignedTx struct {
	ChainTag     uint8               `json:"chainTag"`
	BlockRef     string              `json:"blockRef"`
	Expiration   uint32              `json:"expiration"`
	Clauses      Clauses             `json:"clauses"`
	Gas          uint64              `json:"gas"`
	DependsOn    *polo.Bytes32       `json:"dependsOn"`
	Nonce        math.HexOrDecimal64 `json:"nonce"`
}

func (ustx *UnSignedTx) decode() (*tx.Transaction, error) {
	txBuilder := new(tx.Builder)
	for _, clause := range ustx.Clauses {
		data, err := hexutil.Decode(clause.Data)
		if err != nil {
			return nil, errors.WithMessage(err, "data")
		}
		v := big.Int(clause.Value)
		txBuilder.Clause(tx.NewClause(clause.To).WithData(data).WithValue(&v))
	}
	blockRef, err := hexutil.Decode(ustx.BlockRef)
	if err != nil {
		return nil, errors.WithMessage(err, "blockRef")
	}
	var bf tx.BlockRef
	copy(bf[:], blockRef[:])

	return txBuilder.ChainTag(ustx.ChainTag).
		BlockRef(bf).
		Expiration(ustx.Expiration).
		Gas(ustx.Gas).
		DependsOn(ustx.DependsOn).
		Nonce(uint64(ustx.Nonce)).
		Build(), nil
}

type SignedTx struct {
	UnSignedTx
	Signature string `json:"signature"`
}

func (stx *SignedTx) decode() (*tx.Transaction, error) {
	tx, err := stx.UnSignedTx.decode()
	if err != nil {
		return nil, err
	}
	sig, err := hexutil.Decode(stx.Signature)
	if err != nil {
		return nil, errors.WithMessage(err, "signature")
	}
	return tx.WithSignature(sig), nil
}

type RawTx struct {
	Raw string `json:"raw"`
}

func (rtx *RawTx) Decode() (*tx.Transaction, error) {
	return rtx.decode()
}

func (rtx *RawTx) decode() (*tx.Transaction, error) {
	data, err := hexutil.Decode(rtx.Raw)
	if err != nil {
		return nil, err
	}
	var tx *tx.Transaction
	if err := rlp.DecodeBytes(data, &tx); err != nil {
		return nil, err
	}
	return tx, nil
}

type PlainTx struct {
	Plain string `json:"plain"`
}

func (plaintx *PlainTx) decode() (*tx.Transaction, error) {
	plaindata, err := hexutil.Decode(plaintx.Plain)
	if err != nil {
		return nil, err
	}

	var ptx *plainTransaction
	if err := rlp.DecodeBytes(plaindata, &ptx); err != nil {
		return nil, err
	}

	addr := ptx.To
	if ptx.To.IsZero() {
		addr = nil
	}

	clause := tx.NewClause(addr)
	v := ptx.Value
	clause = clause.WithValue(v)
	clause = clause.WithData(ptx.Data)
	transaction := new(tx.Builder).
		ChainTag(ptx.ChainTag).
		Expiration(120000).
		Gas(ptx.Gas).
		Nonce(ptx.Nonce).
		Clause(clause).
		Build()
	sig, err := hexutil.Decode(ptx.Signature)
	if err != nil {
		return nil, err
	}
	transaction = transaction.WithSignature(sig)
	transaction = transaction.WithPlain(true)
	return transaction, nil
}

type rawTransaction struct {
	RawTx
	Meta TxMeta `json:"meta"`
}

type plainTransaction struct {
	ChainTag  byte   `json:"chaintag"`
	To        *polo.Address `json:"to"`
	Value     *big.Int `json:"value"`
	Data      []byte `json:"data"`
	Gas       uint64 `json:"gas"`
	Nonce     uint64 `json:"nonce"`
	Signature string `json:"signature"`
}

func (ptx *plainTransaction) decode() (*tx.Transaction, error) {

	addr := ptx.To
	if ptx.To.IsZero() {
		addr = nil
	}

	v := ptx.Value
	cla := tx.NewClause(addr).WithValue(v).WithData(ptx.Data)
	transaction := new(tx.Builder).
		Expiration(10).
		Gas(ptx.Gas).
		Nonce(ptx.Nonce).
		Clause(cla).
		Build()
	sig, err := hexutil.Decode(ptx.Signature)
	if err != nil {
		return nil, err
	}
	transaction.WithSignature(sig)
	transaction.WithPlain(true)
	return transaction, nil
}

func convertTransaction(tx *tx.Transaction, header *block.Header, txIndex uint64, clauses bool) (interface{}, error) {
	//tx signer
	signer, err := tx.Signer()
	if err != nil {
		return nil, err
	}
	cls := make(Clauses, len(tx.Clauses()))
	for i, c := range tx.Clauses() {
		cls[i] = convertClause(c)
	}

	if clauses == false {
		t := &PlainTransaction{
			ChainTag: tx.ChainTag(),
			ID:       tx.ID(),
			Origin:   signer,
			Nonce:    math.HexOrDecimal64(tx.Nonce()),
			Gas:      tx.Gas(),
			To:       cls[0].To,
			Value:    cls[0].Value,
			Data:     cls[0].Data,
			Meta: TxMeta{
				BlockID:        header.ID(),
				BlockNumber:    header.Number(),
				BlockTimestamp: header.Timestamp(),
			},
		}

		return t, nil
	} else {
		br := tx.BlockRef()
		t := &Transaction{
			ChainTag:     tx.ChainTag(),
			ID:           tx.ID(),
			Origin:       signer,
			BlockRef:     hexutil.Encode(br[:]),
			Expiration:   tx.Expiration(),
			Nonce:        math.HexOrDecimal64(tx.Nonce()),
			Size:         uint32(tx.Size()),
			Gas:          tx.Gas(),
			DependsOn:    tx.DependsOn(),
			Clauses:      cls,
			Meta: TxMeta{
				BlockID:        header.ID(),
				BlockNumber:    header.Number(),
				BlockTimestamp: header.Timestamp(),
			},
		}
		return t, nil
	}

}

type TxMeta struct {
	BlockID        polo.Bytes32 `json:"blockID"`
	BlockNumber    uint32       `json:"blockNumber"`
	BlockTimestamp uint64       `json:"blockTimestamp"`
}

type LogMeta struct {
	BlockID        polo.Bytes32 `json:"blockID"`
	BlockNumber    uint32       `json:"blockNumber"`
	BlockTimestamp uint64       `json:"blockTimestamp"`
	TxID           polo.Bytes32 `json:"txID"`
	TxOrigin       polo.Address `json:"txOrigin"`
}

//Receipt for json marshal
type Receipt struct {
	GasUsed  uint64                `json:"gasUsed"`
	GasPayer polo.Address          `json:"gasPayer"`
	Paid     *math.HexOrDecimal256 `json:"paid"`
	Reward   *math.HexOrDecimal256 `json:"reward"`
	Reverted bool                  `json:"reverted"`
	Meta     LogMeta               `json:"meta"`
	Outputs  []*Output             `json:"outputs"`
}

// Output output of clause execution.
type Output struct {
	ContractAddress *polo.Address `json:"contractAddress"`
	Events          []*Event      `json:"events"`
	Transfers       []*Transfer   `json:"transfers"`
}

// Event event.
type Event struct {
	Address polo.Address   `json:"address"`
	Topics  []polo.Bytes32 `json:"topics"`
	Data    string         `json:"data"`
}

// Transfer transfer log.
type Transfer struct {
	Sender    polo.Address          `json:"sender"`
	Recipient polo.Address          `json:"recipient"`
	Amount    *math.HexOrDecimal256 `json:"amount"`
}

//ConvertReceipt convert a raw clause into a jason format clause
func ConvertReceipt(txReceipt *tx.Receipt, header *block.Header, tx *tx.Transaction) (*Receipt, error) {
	reward := math.HexOrDecimal256(*txReceipt.Reward)
	paid := math.HexOrDecimal256(*txReceipt.Paid)
	signer, err := tx.Signer()
	if err != nil {
		return nil, err
	}
	receipt := &Receipt{
		GasUsed:  txReceipt.GasUsed,
		GasPayer: txReceipt.GasPayer,
		Paid:     &paid,
		Reward:   &reward,
		Reverted: txReceipt.Reverted,
		Meta: LogMeta{
			header.ID(),
			header.Number(),
			header.Timestamp(),
			tx.ID(),
			signer,
		},
	}
	receipt.Outputs = make([]*Output, len(txReceipt.Outputs))
	for i, output := range txReceipt.Outputs {
		clause := tx.Clauses()[i]
		var contractAddr *polo.Address
		if clause.To() == nil {
			cAddr := polo.CreateContractAddress(tx.ID(), uint32(i), 0)
			contractAddr = &cAddr
		}
		otp := &Output{contractAddr,
			make([]*Event, len(output.Events)),
			make([]*Transfer, len(output.Transfers)),
		}
		for j, txEvent := range output.Events {
			event := &Event{
				Address: txEvent.Address,
				Data:    hexutil.Encode(txEvent.Data),
			}
			event.Topics = make([]polo.Bytes32, len(txEvent.Topics))
			for k, topic := range txEvent.Topics {
				event.Topics[k] = topic
			}
			otp.Events[j] = event

		}
		for j, txTransfer := range output.Transfers {
			transfer := &Transfer{
				Sender:    txTransfer.Sender,
				Recipient: txTransfer.Recipient,
				Amount:    (*math.HexOrDecimal256)(txTransfer.Amount),
			}
			otp.Transfers[j] = transfer
		}
		receipt.Outputs[i] = otp
	}
	return receipt, nil
}
