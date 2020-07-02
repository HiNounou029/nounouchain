// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package polo_test

import (
	"fmt"
	"testing"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/stretchr/testify/assert"
)

func TestBloom(t *testing.T) {

	itemCount := 100
	bloom := polo.NewBloom(polo.EstimateBloomK(itemCount))

	for i := 0; i < itemCount; i++ {
		bloom.Add([]byte(fmt.Sprintf("%v", i)))
	}

	for i := 0; i < itemCount; i++ {
		assert.Equal(t, true, bloom.Test([]byte(fmt.Sprintf("%v", i))))
	}
}
