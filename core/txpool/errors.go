// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package txpool

type (
	badTxError      struct{ msg string }
	txRejectedError struct{ msg string }
)

func (e badTxError) Error() string {
	return "bad tx: " + e.msg
}

func (e txRejectedError) Error() string {
	return "tx rejected: " + e.msg
}

// IsBadTx returns whether the given error indicates that tx is bad.
func IsBadTx(err error) bool {
	_, ok := err.(badTxError)
	return ok
}

// IsTxRejected returns whether the given error indicates tx is rejected.
func IsTxRejected(err error) bool {
	_, ok := err.(txRejectedError)
	return ok
}
