// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package polo_test

import (
	"github.com/HiNounou029/nounouchain/crypto/sha3"
	"testing"
)

func BenchmarkKeccak(b *testing.B) {
	data := []byte("hello world")
	for i := 0; i < b.N; i++ {
		hash := sha3.NewKeccak256()
		hash.Write(data)
		hash.Sum(nil)
	}
}
