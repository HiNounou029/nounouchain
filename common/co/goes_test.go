// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package co_test

import (
	"testing"

	"github.com/HiNounou029/nounouchain/common/co"
)

func TestGoes(t *testing.T) {
	var g co.Goes
	g.Go(func() {})
	g.Go(func() {})
	g.Wait()

	<-g.Done()
}
