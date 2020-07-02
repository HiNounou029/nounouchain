// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package transactions

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/HiNounou029/nounouchain/api/utils"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/core/txpool"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
)

var (
	log = log15.New()
)

type Transactions struct {
	chain *chain.Chain
	pool  *txpool.TxPool
}

func New(chain *chain.Chain, pool *txpool.TxPool) *Transactions {
	return &Transactions{
		chain,
		pool,
	}
}

func (t *Transactions) getRawTransaction(txID polo.Bytes32, blockID polo.Bytes32) (*rawTransaction, error) {
	txMeta, err := t.chain.GetTransactionMeta(txID, blockID)
	if err != nil {
		if t.chain.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	tx, err := t.chain.GetTransaction(txMeta.BlockID, txMeta.Index)
	if err != nil {
		return nil, err
	}
	block, err := t.chain.GetBlock(txMeta.BlockID)
	if err != nil {
		return nil, err
	}
	raw, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}
	return &rawTransaction{
		RawTx: RawTx{hexutil.Encode(raw)},
		Meta: TxMeta{
			BlockID:        block.Header().ID(),
			BlockNumber:    block.Header().Number(),
			BlockTimestamp: block.Header().Timestamp(),
		},
	}, nil
}

func (t *Transactions) getTransactionByID(txID polo.Bytes32, blockID polo.Bytes32, clauses bool) (interface{}, error) {
	txMeta, err := t.chain.GetTransactionMeta(txID, blockID)
	if err != nil {
		if t.chain.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	tx, err := t.chain.GetTransaction(txMeta.BlockID, txMeta.Index)
	if err != nil {
		return nil, err
	}
	h, err := t.chain.GetBlockHeader(txMeta.BlockID)
	if err != nil {
		return nil, err
	}
	return convertTransaction(tx, h, txMeta.Index, clauses)
}

//GetTransactionReceiptByID get tx's receipt
func (t *Transactions) getTransactionReceiptByID(txID polo.Bytes32, blockID polo.Bytes32) (*Receipt, error) {
	txMeta, err := t.chain.GetTransactionMeta(txID, blockID)
	if err != nil {
		if t.chain.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	tx, err := t.chain.GetTransaction(txMeta.BlockID, txMeta.Index)
	if err != nil {
		return nil, err
	}
	h, err := t.chain.GetBlockHeader(txMeta.BlockID)
	if err != nil {
		return nil, err
	}
	receipt, err := t.chain.GetTransactionReceipt(txMeta.BlockID, txMeta.Index)
	if err != nil {
		return nil, err
	}
	return ConvertReceipt(receipt, h, tx)
}
func (t *Transactions) handleSendTransaction(w http.ResponseWriter, req *http.Request) error {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return utils.BadRequest(errors.WithMessage(err, "body"))
	}
	if m == nil {
		return utils.BadRequest(errors.New("body: empty body"))
	}
	var sendTx = func(tx *tx.Transaction) error {
		if err := t.pool.Add(tx); err != nil {
			log.Error("err: ", err)
			if txpool.IsBadTx(err) {
				return utils.BadRequest(err)
			}
			if txpool.IsTxRejected(err) {
				return utils.Forbidden(err)
			}

			return err
		}

		return utils.WriteJSON(w, map[string]string{
			"id": tx.ID().String(),
		})
	}
	reader := bytes.NewReader(data)
	if hasKey(m, "raw") {
		var rawTx *RawTx
		if err := utils.ParseJSON(reader, &rawTx); err != nil {
			return utils.BadRequest(errors.WithMessage(err, "body"))
		}

		tx, err := rawTx.decode()
		if err != nil {
			log.Error("rawTx", "err", err)
			return utils.BadRequest(errors.WithMessage(err, "raw"))
		}
		return sendTx(tx)
	} else if hasKey(m, "signature") {
		var stx *SignedTx
		if err := utils.ParseJSON(reader, &stx); err != nil {
			return utils.BadRequest(errors.WithMessage(err, "body"))
		}
		tx, err := stx.decode()
		if err != nil {
			return utils.BadRequest(err)
		}
		return sendTx(tx)
	} else if hasKey(m, "plain") {
		var ptx *PlainTx
		if err := utils.ParseJSON(reader, &ptx); err != nil {
			return utils.BadRequest(errors.WithMessage(err, "body"))
		}

		tx, err := ptx.decode()
		if err != nil {
			return utils.BadRequest(err)
		}
		return sendTx(tx)
	} else {
		var ustx *UnSignedTx
		if err := utils.ParseJSON(reader, &ustx); err != nil {
			return utils.BadRequest(errors.WithMessage(err, "body"))
		}
		tx, err := ustx.decode()
		if err != nil {
			return utils.BadRequest(err)
		}
		return utils.WriteJSON(w, map[string]string{
			"signingHash": tx.SigningHash().String(),
		})
	}
}

func (t *Transactions) handleGetTransactionByID(w http.ResponseWriter, req *http.Request) error {
	id := mux.Vars(req)["id"]
	txID, err := polo.ParseBytes32(id)
	if err != nil {
		return utils.BadRequest(errors.WithMessage(err, "id"))
	}
	head, err := t.parseHead(req.URL.Query().Get("head"))
	if err != nil {
		return utils.BadRequest(errors.WithMessage(err, "head"))
	}
	h, err := t.chain.GetBlockHeader(head)
	if err != nil {
		if t.chain.IsNotFound(err) {
			return utils.BadRequest(errors.WithMessage(err, "head"))
		}
		return err
	}
	raw := req.URL.Query().Get("raw")
	if raw != "" && raw != "false" && raw != "true" {
		return utils.BadRequest(errors.WithMessage(errors.New("should be boolean"), "raw"))
	}
	if raw == "true" {
		tx, err := t.getRawTransaction(txID, h.ID())
		if err != nil {
			return err
		}
		return utils.WriteJSON(w, tx)
	}

	//clauses := req.URL.Query().Get("clauses")
	clauses := req.URL.Query().Get("clauses")
	tx, err := t.getTransactionByID(txID, h.ID(), (clauses == "true"))
	if err != nil {
		return err
	}

	return utils.WriteTo(w, req, tx)
}

func (t *Transactions) handleGetTransactionReceiptByID(w http.ResponseWriter, req *http.Request) error {
	id := mux.Vars(req)["id"]
	txID, err := polo.ParseBytes32(id)
	if err != nil {
		return utils.BadRequest(errors.WithMessage(err, "id"))
	}
	head, err := t.parseHead(req.URL.Query().Get("head"))
	if err != nil {
		return utils.BadRequest(errors.WithMessage(err, "head"))
	}
	h, err := t.chain.GetBlockHeader(head)
	if err != nil {
		if t.chain.IsNotFound(err) {
			return utils.BadRequest(errors.WithMessage(err, "head"))
		}
		return err
	}
	receipt, err := t.getTransactionReceiptByID(txID, h.ID())
	if err != nil {
		return err
	}
	return utils.WriteTo(w, req, receipt)
}

func (t *Transactions) parseHead(head string) (polo.Bytes32, error) {
	if head == "" {
		return t.chain.BestBlock().Header().ID(), nil
	}
	h, err := polo.ParseBytes32(head)
	if err != nil {
		return polo.Bytes32{}, err
	}
	return h, nil
}

func (t *Transactions) Mount(root *mux.Router, pathPrefix string) {
	sub := root.PathPrefix(pathPrefix).Subrouter()

	sub.Path("").Methods("POST").HandlerFunc(utils.WrapHandlerFunc(t.handleSendTransaction))
	sub.Path("/{id}").Methods("GET").HandlerFunc(utils.WrapHandlerFunc(t.handleGetTransactionByID))
	sub.Path("/{id}/receipt").Methods("GET").HandlerFunc(utils.WrapHandlerFunc(t.handleGetTransactionReceiptByID))
}
