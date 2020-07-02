// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package node

import (
	"crypto/ecdsa"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/crypto"
)

type Master struct {
	PrivateKey  *ecdsa.PrivateKey
	Beneficiary *polo.Address
}

func (m *Master) Address() polo.Address {
	return polo.Address(crypto.PubkeyToAddress(m.PrivateKey.PublicKey))
}
