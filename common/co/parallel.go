// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package co

import (
	"runtime"
	"sync/atomic"
)

var numCPU = runtime.NumCPU()

// Parallel to run a batch of work using as many CPU as it can.
func Parallel(cb func(chan<- func())) <-chan struct{} {
	queue := make(chan func(), numCPU*16)
	defer close(queue)

	done := make(chan struct{})

	nGo := int32(numCPU)
	for i := 0; i < numCPU; i++ {
		go func() {
			for work := range queue {
				work()
			}

			if atomic.AddInt32(&nGo, -1) == 0 {
				close(done)
			}
		}()
	}
	cb(queue)
	return done
}
