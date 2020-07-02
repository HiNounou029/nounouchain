// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package runtime

import (
	"fmt"
	"math/big"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/builtin"
	"github.com/HiNounou029/nounouchain/common/xenv"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/pkg/errors"
)

// ResolvedTransaction resolve the transaction according to given state.
type ResolvedTransaction struct {
	tx           *tx.Transaction
	Origin       polo.Address
	IntrinsicGas uint64
	Clauses      []*tx.Clause
}

// ResolveTransaction resolves the transaction and performs basic validation.
func ResolveTransaction(tx *tx.Transaction) (*ResolvedTransaction, error) {
	origin, err := tx.Signer()
	if err != nil {
		return nil, err
	}
	intrinsicGas, err := tx.IntrinsicGas()
	if err != nil {
		return nil, err
	}
	if tx.Gas() < intrinsicGas {
		return nil, fmt.Errorf("address: %s, intrinsic gas exceeds provided gas want %v, have %v", origin.String(),intrinsicGas, tx.Gas())
	}

	clauses := tx.Clauses()
	sumValue := new(big.Int)
	for _, clause := range clauses {
		value := clause.Value()
		if value.Sign() < 0 {
			return nil, errors.New("clause with negative value")
		}

		sumValue.Add(sumValue, value)
		if sumValue.Cmp(math.MaxBig256) > 0 {
			return nil, errors.New("tx value too large")
		}
	}

	return &ResolvedTransaction{
		tx,
		origin,
		intrinsicGas,
		clauses,
	}, nil
}

// CommonTo returns common 'To' field of clauses if any.
// Nil returned if no common 'To'.
func (r *ResolvedTransaction) CommonTo() *polo.Address {
	if len(r.Clauses) == 0 {
		return nil
	}

	firstTo := r.Clauses[0].To()
	if firstTo == nil {
		return nil
	}

	for _, clause := range r.Clauses[1:] {
		to := clause.To()
		if to == nil {
			return nil
		}
		if *to != *firstTo {
			return nil
		}
	}
	return firstTo
}

// BuyGas consumes balance to buy gas, to prepare for execution.
func (r *ResolvedTransaction) BuyGas(state *state.State) (
	gasPrice *big.Int,
	returnGas func(uint64), err error) {
	gasPrice = builtin.Params.Native(state).Get(polo.KeyBaseGasPrice)
	doReturnGas := func(rgas uint64) *big.Int {
		returnedGas := new(big.Int).Mul(new(big.Int).SetUint64(rgas), gasPrice)
		state.AddBalance(r.Origin, returnedGas)
		return returnedGas
	}
	prepaid := new(big.Int).Mul(new(big.Int).SetUint64(r.tx.Gas()), gasPrice)
	if state.SubBalance(r.Origin, prepaid) {
		return gasPrice, func(rgas uint64) { doReturnGas(rgas) }, nil
	}
	return nil, nil, fmt.Errorf("insufficient gas, require at least %d, addr: %s", prepaid, r.Origin.String())
}

// ToContext create a tx context object.
func (r *ResolvedTransaction) ToContext(gasPrice *big.Int, blockNumber uint32, getID func(uint32) polo.Bytes32) *xenv.TransactionContext {
	return &xenv.TransactionContext{
		ID:         r.tx.ID(),
		Origin:     r.Origin,
		GasPrice:   gasPrice,
		BlockRef:   r.tx.BlockRef(),
		Expiration: r.tx.Expiration(),
	}
}
