// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package tx

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/common/metric"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/inconshreveable/log15"
)

var log = log15.New("pkg", "tx")

var (
	errIntrinsicGasOverflow = errors.New("intrinsic gas overflow")
)

// Transaction is an immutable tx type.
type Transaction struct {
	body body

	cache struct {
		signingHash  atomic.Value
		signer       atomic.Value
		id           atomic.Value
		unprovedWork atomic.Value
		size         atomic.Value
		intrinsicGas atomic.Value
	}
}

// body describes details of a tx.
type body struct {
	ChainTag   byte
	BlockRef   uint64
	Expiration uint32
	Clauses    []*Clause
	Gas        uint64
	DependsOn  *polo.Bytes32 `rlp:"nil"`
	Nonce      uint64
	Reserved   []interface{}
	Signature  []byte
	Plain      bool
}

// ChainTag returns chain tag.
func (t *Transaction) ChainTag() byte {
	return t.body.ChainTag
}

// Nonce returns nonce value.
func (t *Transaction) Nonce() uint64 {
	return t.body.Nonce
}

// BlockRef returns block reference, which is first 8 bytes of block hash.
func (t *Transaction) BlockRef() (br BlockRef) {
	binary.BigEndian.PutUint64(br[:], t.body.BlockRef)
	return
}

// Expiration returns expiration in unit block.
// A valid transaction requires:
// blockNum in [blockRef.Num... blockRef.Num + Expiration]
func (t *Transaction) Expiration() uint32 {
	return t.body.Expiration
}

// IsExpired returns whether the tx is expired according to the given blockNum.
func (t *Transaction) IsExpired(blockNum uint32) bool {
	if t.BlockRef().Number() <= 0 {
		return false //如果用户不指定block ref,则默认不过期
	}
	return uint64(blockNum) > uint64(t.BlockRef().Number())+uint64(t.body.Expiration) // cast to uint64 to prevent potential overflow
}

// ID returns id of tx.
// ID = hash(signingHash, signer).
// It returns zero Bytes32 if signer not available.
func (t *Transaction) ID() (id polo.Bytes32) {
	if cached := t.cache.id.Load(); cached != nil {
		return cached.(polo.Bytes32)
	}
	defer func() { t.cache.id.Store(id) }()

	signer, err := t.Signer()
	if err != nil {
		return
	}
	hw := polo.NewBlake2b()
	hw.Write(t.SigningHash().Bytes())
	hw.Write(signer.Bytes())
	hw.Sum(id[:0])
	return
}

// SigningHash returns hash of tx excludes signature.
func (t *Transaction) SigningHash() (hash polo.Bytes32) {
	if cached := t.cache.signingHash.Load(); cached != nil {
		return cached.(polo.Bytes32)
	}
	defer func() { t.cache.signingHash.Store(hash) }()

	hw := polo.NewBlake2b()
	if t.body.Plain {
		value := t.Clauses()[0].Value()
		to := t.Clauses()[0].To()
		if to == nil {
			address := polo.BytesToAddress([]byte(""))
			to = &address
		}
		rlp.Encode(hw, []interface{}{
			t.body.ChainTag,
			to,
			value,
			t.Clauses()[0].Data(),
			t.body.Gas,
			t.body.Nonce,
		})

		//fmt.Printf("%v\n%v\n%v\n%v\n%v\n%v\n", t.body.ChainTag, to, value, t.Clauses()[0].Data(), t.body.Gas, t.body.Nonce)
	} else {
		rlp.Encode(hw, []interface{}{
			t.body.ChainTag,
			t.body.BlockRef,
			t.body.Expiration,
			t.body.Clauses,
			t.body.Gas,
			t.body.DependsOn,
			t.body.Nonce,
			t.body.Reserved,
		})
	}
	hw.Sum(hash[:0])
	return
}

// Gas returns gas provision for this tx.
func (t *Transaction) Gas() uint64 {
	return t.body.Gas
}

// Clauses returns caluses in tx.
func (t *Transaction) Clauses() []*Clause {
	return append([]*Clause(nil), t.body.Clauses...)
}

// DependsOn returns depended tx hash.
func (t *Transaction) DependsOn() *polo.Bytes32 {
	if t.body.DependsOn == nil {
		return nil
	}
	cpy := *t.body.DependsOn
	return &cpy
}

// Signature returns signature.
func (t *Transaction) Signature() []byte {
	return append([]byte(nil), t.body.Signature...)
}

// Signer extract signer of tx from signature.
func (t *Transaction) Signer() (signer polo.Address, err error) {
	if cached := t.cache.signer.Load(); cached != nil {
		return cached.(polo.Address), nil
	}
	defer func() {
		if err == nil {
			t.cache.signer.Store(signer)
		}
	}()

	//	hash := t.SigningHash().Bytes()
	//	fmt.Println("hash: ", hash)
	pub, err := crypto.SigToPub(t.SigningHash().Bytes(), t.body.Signature)
	if err != nil {
		return polo.Address{}, err
	}
	signer = polo.Address(crypto.PubkeyToAddress(*pub))
	//	fmt.Println("sig: ", signer)
	return
}

func (t *Transaction) Plain() bool {
	return t.body.Plain
}

// WithSignature create a new tx with signature set.
func (t *Transaction) WithSignature(sig []byte) *Transaction {
	newTx := Transaction{
		body: t.body,
	}
	// copy sig
	newTx.body.Signature = append([]byte(nil), sig...)
	return &newTx
}

func (t *Transaction) WithPlain(plain bool) *Transaction {
	newTx := Transaction{
		body: t.body,
	}
	newTx.body.Plain = plain
	return &newTx
}

// HasReservedFields returns if there're reserved fields.
// Reserved fields are for backward compatibility purpose.
func (t *Transaction) HasReservedFields() bool {
	return len(t.body.Reserved) > 0
}

// EncodeRLP implements rlp.Encoder
func (t *Transaction) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &t.body)
}

// DecodeRLP implements rlp.Decoder
func (t *Transaction) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	var body body
	if err := s.Decode(&body); err != nil {
		return err
	}
	*t = Transaction{body: body}

	t.cache.size.Store(metric.StorageSize(rlp.ListSize(size)))
	return nil
}

// Size returns size in bytes when RLP encoded.
func (t *Transaction) Size() metric.StorageSize {
	if cached := t.cache.size.Load(); cached != nil {
		return cached.(metric.StorageSize)
	}
	var size metric.StorageSize
	rlp.Encode(&size, t)
	t.cache.size.Store(size)
	return size
}

// IntrinsicGas returns intrinsic gas of tx.
func (t *Transaction) IntrinsicGas() (uint64, error) {
	if cached := t.cache.intrinsicGas.Load(); cached != nil {
		return cached.(uint64), nil
	}

	gas, err := IntrinsicGas(t.body.Clauses...)
	if err != nil {
		return 0, err
	}
	t.cache.intrinsicGas.Store(gas)
	return gas, nil
}

func (t *Transaction) String() string {
	var (
		from      string
		br        BlockRef
		dependsOn string
	)
	signer, err := t.Signer()
	if err != nil {
		from = "N/A"
	} else {
		from = signer.String()
	}

	binary.BigEndian.PutUint64(br[:], t.body.BlockRef)
	if t.body.DependsOn == nil {
		dependsOn = "nil"
	} else {
		dependsOn = t.body.DependsOn.String()
	}

	return fmt.Sprintf(`
	Tx(%v, %v)
	From:           %v
	Clauses:        %v
	Gas:            %v
	ChainTag:       %v
	BlockRef:       %v-%x
	Expiration:     %v
	DependsOn:      %v
	Nonce:          %v
	Signature:      0x%x
`, t.ID(), t.Size(), from, t.body.Clauses, t.body.Gas,
		t.body.ChainTag, br.Number(), br[4:], t.body.Expiration, dependsOn, t.body.Nonce, t.body.Signature)
}

// IntrinsicGas calculate intrinsic gas cost for tx with such clauses.
func IntrinsicGas(clauses ...*Clause) (uint64, error) {
	if len(clauses) == 0 {
		return polo.TxGas + polo.ClauseGas, nil
	}

	var total = polo.TxGas
	var overflow bool
	for _, c := range clauses {
		gas, err := dataGas(c.body.Data)
		if err != nil {
			return 0, err
		}
		total, overflow = math.SafeAdd(total, gas)
		if overflow {
			return 0, errIntrinsicGasOverflow
		}

		var cgas uint64
		if c.IsCreatingContract() {
			// contract creation
			cgas = polo.ClauseGasContractCreation
		} else {
			cgas = polo.ClauseGas
		}

		total, overflow = math.SafeAdd(total, cgas)
		if overflow {
			return 0, errIntrinsicGasOverflow
		}
	}
	return total, nil
}

// see core.IntrinsicGas
func dataGas(data []byte) (uint64, error) {
	if len(data) == 0 {
		return 0, nil
	}
	var z, nz uint64
	for _, byt := range data {
		if byt == 0 {
			z++
		} else {
			nz++
		}
	}
	zgas, overflow := math.SafeMul(params.TxDataZeroGas, z)
	if overflow {
		return 0, errIntrinsicGasOverflow
	}
	nzgas, overflow := math.SafeMul(params.TxDataNonZeroGas, nz)
	if overflow {
		return 0, errIntrinsicGasOverflow
	}

	gas, overflow := math.SafeAdd(zgas, nzgas)
	if overflow {
		return 0, errIntrinsicGasOverflow
	}
	return gas, nil
}
