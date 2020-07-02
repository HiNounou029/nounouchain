// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package poloclient

import (
	"encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/mclock"
)

const (
	// HashLength is the expected length of the hash
	HashLength = 32
	// AddressLength is the expected length of the adddress
	AddressLength = 20
)

type plainTransaction struct {
	ChainTag  byte   `json:"chaintag"`
	To        string `json:"to"`
	Value     uint64 `json:"value"`
	Data      string `json:"data"`
	Gas       uint64 `json:"gas"`
	Nonce     uint64 `json:"nonce"`
	Signature string `json:"signature"`
}

func NewPlainTransaction(client *PoloClient, addr *Address, value uint64, data []byte) *plainTransaction {
	var plain plainTransaction
	plain.ChainTag = client.ChainStatus.Tag
	if addr != nil {
		plain.To = addr.String()
	} else {
		address := BytesToAddress([]byte(""))
		plain.To = address.String()
	}

	plain.Value = value

	if data != nil {
		plain.Data = hexutil.Encode([]byte(data))
	} else {
		plain.Data = hexutil.Encode([]byte(nil))
	}

	plain.Gas = uint64(1000000)
	plain.Nonce = uint64(mclock.Now())
	return &plain
}

func NewBlockRef(blockNum uint32) (br BlockRef) {
	binary.BigEndian.PutUint32(br[:], blockNum)
	return
}

func (t *plainTransaction) SigningHash() (hash Bytes32) {
	hw := NewBlake2b()
	rawString := fmt.Sprintf("%v\n%v\n%v\n%v\n%v",
		t.To,
		t.Value,
		t.Data,
		t.Gas,
		t.Nonce)
	hw.Write([]byte(rawString))
	hw.Sum(hash[:0])
	return
}

func (t *plainTransaction) WithSignature(sig []byte) *plainTransaction {
	newTx := t
	// copy sig
	newTx.Signature = hexutil.Encode(sig)
	return newTx
}

type RawTx struct {
	Raw string `json:"raw"`
}

type PlainTx struct {
	Plain string `json:"plain"`
}
