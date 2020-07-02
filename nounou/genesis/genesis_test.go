// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package genesis_test

import (
	"testing"

	"github.com/HiNounou029/nounouchain/nounou/genesis"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/stretchr/testify/assert"
)

func TestTestnetGenesis(t *testing.T) {
	kv, _ := storage.NewMem()
	gene := genesis.NewDevnet()

	b0, _, err := gene.Build(state.NewCreator(kv))
	assert.Nil(t, err)

	_, err = state.New(b0.Header().StateRoot(), kv)
	assert.Nil(t, err)
}

func TestTime( t *testing.T) {
	//gt := time.Date(2018, 11, 22, 0, 0, 0, 0, time.Local)
	//fmt.Printf("%s\n", gt.String())
	//
	//fmt.Printf("%d\n", gt.Unix())
}
