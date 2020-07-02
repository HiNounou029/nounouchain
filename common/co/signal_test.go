// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package co_test

import (
	"testing"

	"github.com/HiNounou029/nounouchain/common/co"
	"github.com/stretchr/testify/assert"
)

func TestSignal_SignalBeforeWait(t *testing.T) {
	var sig co.Signal
	sig.Signal()

	<-sig.NewWaiter().C()
}

func TestSignal_SignalAfterWait(t *testing.T) {
	var sig co.Signal
	w := sig.NewWaiter()
	sig.Signal()
	<-w.C()
}

func TestSignal_BroadcastBefore(t *testing.T) {
	var sig co.Signal
	sig.Broadcast()

	var ws []co.Waiter
	for i := 0; i < 10; i++ {
		ws = append(ws, sig.NewWaiter())
	}

	var n int
	for _, w := range ws {
		select {
		case <-w.C():
		default:
			n++
		}
	}
	assert.Equal(t, 10, n)
}

func TestSignal_BroadcastAfterWait(t *testing.T) {
	var sig co.Signal

	var ws []co.Waiter
	for i := 0; i < 10; i++ {
		ws = append(ws, sig.NewWaiter())
	}

	sig.Broadcast()

	for _, w := range ws {
		<-w.C()
	}
}
