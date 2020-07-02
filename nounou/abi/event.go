// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package abi

import (
	"github.com/HiNounou029/nounouchain/polo"
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

// Event see abi.Event in go-ethereum.
type Event struct {
	id                 polo.Bytes32
	event              *ethabi.Event
	argsWithoutIndexed ethabi.Arguments
}

func newEvent(event *ethabi.Event) *Event {
	var argsWithoutIndexed ethabi.Arguments
	for _, arg := range event.Inputs {
		if !arg.Indexed {
			argsWithoutIndexed = append(argsWithoutIndexed, arg)
		}
	}
	return &Event{
		polo.Bytes32(event.Id()),
		event,
		argsWithoutIndexed,
	}
}

// ID returns event id.
func (e *Event) ID() polo.Bytes32 {
	//fmt.Println("event id: ", common.Bytes2Hex(e.id.Bytes()))
	return e.id
}

// Name returns event name.
func (e *Event) Name() string {
	return e.event.Name
}

// Encode encodes args to data.
func (e *Event) Encode(args ...interface{}) ([]byte, error) {
	return e.argsWithoutIndexed.Pack(args...)
}

// Decode decodes event data.
func (e *Event) Decode(data []byte, v interface{}) error {
	return e.argsWithoutIndexed.Unpack(v, data)
}
