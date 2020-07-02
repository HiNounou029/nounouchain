// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package txpool

import (
	"math/big"
	"sort"
	"time"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/vm/runtime"
	"github.com/pkg/errors"
)

type txObject struct {
	*tx.Transaction
	resolved *runtime.ResolvedTransaction

	timeAdded       int64
	executable      bool
	overallGasPrice *big.Int // don't touch this value, it's only be used in pool's housekeeping
}

func resolveTx(tx *tx.Transaction) (*txObject, error) {
	resolved, err := runtime.ResolveTransaction(tx)
	if err != nil {
		return nil, err
	}

	return &txObject{
		Transaction: tx,
		resolved:    resolved,
		timeAdded:   time.Now().UnixNano(),
	}, nil
}

func (o *txObject) Origin() polo.Address {
	return o.resolved.Origin
}

func (o *txObject) Executable(chain *chain.Chain, state *state.State, headBlock *block.Header) (bool, error) {
	switch {
	case o.Gas() > headBlock.GasLimit():
		return false, errors.New("gas too large")
	case o.IsExpired(headBlock.Number()):
		return false, errors.New("expired")
	case o.BlockRef().Number() > headBlock.Number()+uint32(3600*24/polo.Conf.BlockInterval):
		return false, errors.New("block ref out of schedule")
	}

	if _, err := chain.GetTransactionMeta(o.ID(), headBlock.ID()); err != nil {
		if !chain.IsNotFound(err) {
			return false, err
		}
	} else {
		return false, errors.New("known tx")
	}

	if dep := o.DependsOn(); dep != nil {
		txMeta, err := chain.GetTransactionMeta(*dep, headBlock.ID())
		if err != nil {
			if chain.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		if txMeta.Reverted {
			return false, errors.New("dep reverted")
		}
	}

	if o.BlockRef().Number() > headBlock.Number() {
		return false, nil
	}

	checkpoint := state.NewCheckpoint()
	defer state.RevertTo(checkpoint)

	if _, _, err := o.resolved.BuyGas(state); err != nil {
		return false, err
	}
	return true, nil
}

func sortTxObjsByOverallGasPriceDesc(txObjs []*txObject) {
	sort.Slice(txObjs, func(i, j int) bool {
		gp1, gp2 := txObjs[i].overallGasPrice, txObjs[j].overallGasPrice
		return gp1.Cmp(gp2) >= 0
	})
}
