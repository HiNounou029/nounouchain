// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package tx_test

import (
	"fmt"
	"testing"

	. "github.com/HiNounou029/nounouchain/core/tx"
)

func TestReceipt(t *testing.T) {
	var rs Receipts
	fmt.Println(rs.RootHash())

	var txs Transactions
	fmt.Println(txs.RootHash())
}
