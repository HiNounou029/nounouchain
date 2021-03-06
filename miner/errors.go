// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package miner

import "github.com/pkg/errors"

var (
	errGasLimitReached       = errors.New("gas limit reached")
	errTxNotAdoptableNow     = errors.New("tx not adoptable now")
	errTxNotAdoptableForever = errors.New("tx not adoptable forever")
	errKnownTx               = errors.New("known tx")
)

// IsGasLimitReached block if full of txs.
func IsGasLimitReached(err error) bool {
	return errors.Cause(err) == errGasLimitReached
}

// IsTxNotAdoptableNow tx can not be adopted now.
func IsTxNotAdoptableNow(err error) bool {
	return errors.Cause(err) == errTxNotAdoptableNow
}

// IsBadTx not a valid tx.
func IsBadTx(err error) bool {
	_, ok := errors.Cause(err).(badTxError)
	return ok
}

// IsKnownTx tx is already adopted, or in the chain.
func IsKnownTx(err error) bool {
	return errors.Cause(err) == errKnownTx
}

type badTxError struct {
	msg string
}

func (e badTxError) Error() string {
	return "bad tx: " + e.msg
}
