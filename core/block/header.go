// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

//+build !gm

package block

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

// Header contains almost all information about a block, except block body.
// It's immutable.
type Header struct {
	body headerBody

	cache struct {
		signingHash atomic.Value
		signer      atomic.Value
		id          atomic.Value
	}
}

// headerBody body of header
type headerBody struct {
	ParentID    polo.Bytes32
	Timestamp   uint64
	GasLimit    uint64
	Beneficiary polo.Address

	GasUsed    uint64
	TotalScore uint64

	TxsRoot      polo.Bytes32
	StateRoot    polo.Bytes32
	ReceiptsRoot polo.Bytes32

	Signature []byte
}

// ParentID returns id of parent block.
func (h *Header) ParentID() polo.Bytes32 {
	return h.body.ParentID
}

// Number returns sequential number of this block.
func (h *Header) Number() uint32 {
	// inferred from parent id
	return Number(h.body.ParentID) + 1
}

// Timestamp returns timestamp of this block.
func (h *Header) Timestamp() uint64 {
	return h.body.Timestamp
}

// TotalScore returns total score that cumulated from genesis block to this one.
func (h *Header) TotalScore() uint64 {
	return h.body.TotalScore
}

// GasLimit returns gas limit of this block.
func (h *Header) GasLimit() uint64 {
	return h.body.GasLimit
}

// GasUsed returns gas used by txs.
func (h *Header) GasUsed() uint64 {
	return h.body.GasUsed
}

// Beneficiary returns reward recipient.
func (h *Header) Beneficiary() polo.Address {
	return h.body.Beneficiary
}

// TxsRoot returns merkle root of txs contained in this block.
func (h *Header) TxsRoot() polo.Bytes32 {
	return h.body.TxsRoot
}

// StateRoot returns account state merkle root just afert this block being applied.
func (h *Header) StateRoot() polo.Bytes32 {
	return h.body.StateRoot
}

// ReceiptsRoot returns merkle root of tx receipts.
func (h *Header) ReceiptsRoot() polo.Bytes32 {
	return h.body.ReceiptsRoot
}

// ID computes id of block.
// The block ID is defined as: blockNumber + hash(signingHash, signer)[4:].
func (h *Header) ID() (id polo.Bytes32) {
	if cached := h.cache.id.Load(); cached != nil {
		return cached.(polo.Bytes32)
	}
	defer func() {
		// overwrite first 4 bytes of block hash to block number.
		binary.BigEndian.PutUint32(id[:], h.Number())
		h.cache.id.Store(id)
	}()

	signer, err := h.Signer()
	if err != nil {
		return
	}

	hw := polo.NewBlake2b()
	hw.Write(h.SigningHash().Bytes())
	hw.Write(signer.Bytes())
	hw.Sum(id[:0])

	return
}

// SigningHash computes hash of all header fields excluding signature.
func (h *Header) SigningHash() (hash polo.Bytes32) {
	if cached := h.cache.signingHash.Load(); cached != nil {
		return cached.(polo.Bytes32)
	}
	defer func() { h.cache.signingHash.Store(hash) }()

	hw := polo.NewBlake2b()
	rlp.Encode(hw, []interface{}{
		h.body.ParentID,
		h.body.Timestamp,
		h.body.GasLimit,
		h.body.Beneficiary,

		h.body.GasUsed,
		h.body.TotalScore,

		h.body.TxsRoot,
		h.body.StateRoot,
		h.body.ReceiptsRoot,
	})
	hw.Sum(hash[:0])
	return
}

// Signature returns signature.
func (h *Header) Signature() []byte {
	return append([]byte(nil), h.body.Signature...)
}

// withSignature create a new Header object with signature set.
func (h *Header) withSignature(sig []byte) *Header {
	cpy := Header{body: h.body}
	cpy.body.Signature = append([]byte(nil), sig...)
	return &cpy
}

// Signer extract signer of the block from signature.
func (h *Header) Signer() (signer polo.Address, err error) {
	if h.Number() == 0 {
		// special case for genesis block
		return polo.Address{}, nil
	}

	if cached := h.cache.signer.Load(); cached != nil {
		return cached.(polo.Address), nil
	}
	defer func() {
		if err == nil {
			h.cache.signer.Store(signer)
		}
	}()

	pub, err := crypto.SigToPub(h.SigningHash().Bytes(), h.body.Signature)
	if err != nil {
		return polo.Address{}, err
	}

	signer = polo.Address(crypto.PubkeyToAddress(*pub))
	return
}

// EncodeRLP implements rlp.Encoder
func (h *Header) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &h.body)
}

// DecodeRLP implements rlp.Decoder.
func (h *Header) DecodeRLP(s *rlp.Stream) error {
	var body headerBody

	if err := s.Decode(&body); err != nil {
		return err
	}
	*h = Header{body: body}
	return nil
}

func (h *Header) String() string {
	var signerStr string
	if signer, err := h.Signer(); err != nil {
		signerStr = "N/A"
	} else {
		signerStr = signer.String()
	}

	return fmt.Sprintf(`Header(%v):
	Number:			%v
	ParentID:		%v
	Timestamp:		%v
	Signer:			%v
	Beneficiary:	%v
	GasLimit:		%v
	GasUsed:		%v
	TotalScore:		%v
	TxsRoot:		%v
	StateRoot:		%v
	ReceiptsRoot:	%v
	Signature:		0x%x`, h.ID(), h.Number(), h.body.ParentID, h.body.Timestamp, signerStr,
		h.body.Beneficiary, h.body.GasLimit, h.body.GasUsed, h.body.TotalScore,
		h.body.TxsRoot, h.body.StateRoot, h.body.ReceiptsRoot, h.body.Signature)
}

// Number extract block number from block id.
func Number(blockID polo.Bytes32) uint32 {
	// first 4 bytes are over written by block number (big endian).
	return binary.BigEndian.Uint32(blockID[:])
}
