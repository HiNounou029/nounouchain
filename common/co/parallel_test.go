// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package co_test

import (
	"testing"
	"time"

	"github.com/HiNounou029/nounouchain/common/co"
)

func TestParallel(t *testing.T) {
	n := 50
	fn := func() {
		time.Sleep(time.Millisecond * 20)
	}

	startTime := time.Now().UnixNano()
	for i := 0; i < n; i++ {
		fn()
	}
	t.Log("non-parallel", time.Duration(time.Now().UnixNano()-startTime))

	startTime = time.Now().UnixNano()
	<-co.Parallel(func(queue chan<- func()) {
		for i := 0; i < n; i++ {
			queue <- fn
		}
	})
	t.Log("parallel", time.Duration(time.Now().UnixNano()-startTime))
}
