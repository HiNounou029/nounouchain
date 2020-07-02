/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package idemix

import (
	"crypto/ecdsa"

	"github.com/HiNounou029/nounouchain/caclient/amcl"		// mod by syl 2018-08-25=====
	fp256bn "github.com/HiNounou029/nounouchain/caclient/amcl/FP256BN"			// mod by syl 2018-08-25=====
	"github.com/HiNounou029/nounouchain/caclient/idemix"		// mod by syl 2018-08-25=======
)

// Lib represents idemix library
type Lib interface {
	NewIssuerKey(AttributeNames []string, rng *amcl.RAND) (*idemix.IssuerKey, error)
	NewCredential(key *idemix.IssuerKey, m *idemix.CredRequest, attrs []*fp256bn.BIG, rng *amcl.RAND) (*idemix.Credential, error)
	CreateCRI(key *ecdsa.PrivateKey, unrevokedHandles []*fp256bn.BIG, epoch int, alg idemix.RevocationAlgorithm, rng *amcl.RAND) (*idemix.CredentialRevocationInformation, error)
	GenerateLongTermRevocationKey() (*ecdsa.PrivateKey, error)
	GetRand() (*amcl.RAND, error)
	RandModOrder(rng *amcl.RAND) *fp256bn.BIG
}

// libImpl is adapter for idemix library. It implements Lib interface
type libImpl struct{}

// NewLib returns an instance of an object that implements Lib interface
func NewLib() Lib {
	return &libImpl{}
}

func (i *libImpl) GetRand() (*amcl.RAND, error) {
	return idemix.GetRand()
}
func (i *libImpl) NewCredential(key *idemix.IssuerKey, m *idemix.CredRequest, attrs []*fp256bn.BIG, rng *amcl.RAND) (*idemix.Credential, error) {
	return idemix.NewCredential(key, m, attrs, rng)
}
func (i *libImpl) RandModOrder(rng *amcl.RAND) *fp256bn.BIG {
	return idemix.RandModOrder(rng)
}
func (i *libImpl) NewIssuerKey(AttributeNames []string, rng *amcl.RAND) (*idemix.IssuerKey, error) {
	return idemix.NewIssuerKey(AttributeNames, rng)
}
func (i *libImpl) CreateCRI(key *ecdsa.PrivateKey, unrevokedHandles []*fp256bn.BIG, epoch int, alg idemix.RevocationAlgorithm, rng *amcl.RAND) (*idemix.CredentialRevocationInformation, error) {
	return idemix.CreateCRI(key, unrevokedHandles, epoch, alg, rng)
}
func (i *libImpl) GenerateLongTermRevocationKey() (*ecdsa.PrivateKey, error) {
	return idemix.GenerateLongTermRevocationKey()
}
