// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

// +build gm

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tjfoc/gmsm/sm2"
	"math/big"
)

const (
	SigLen = 130 // sm2 signature appending with public key
)

// Ecrecover returns the uncompressed public key that created the given signature.
func Ecrecover(hash, sig []byte) ([]byte, error) {
	byteLen := (sm2.P256Sm2().Params().BitSize + 7) >> 3
	if len(sig) <= 2*byteLen+1 {
		return nil, fmt.Errorf("signature size is smaller than %v", 2*byteLen+1)
	}
	return sig[len(sig)-1-2*byteLen:], nil
}

// SigToPub returns the public key that created the given signature.
func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	s, err := Ecrecover(hash, sig)
	if err != nil {
		return nil, err
	}

	x, y := elliptic.Unmarshal(S256(), s)
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

// Sign calculates an sm2 signature.
// The produced signature is in the [R || S || V] format where V is 1.
func Sign(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}
	var sm2_prv sm2.PrivateKey
	sm2_prv.D = prv.D
	sm2_prv.PublicKey.X = prv.PublicKey.X
	sm2_prv.PublicKey.Y = prv.PublicKey.Y
	sm2_prv.PublicKey.Curve = prv.PublicKey.Curve

	pubData := elliptic.Marshal(S256(), prv.PublicKey.X, prv.PublicKey.Y)
	rawSig, err := sm2_prv.Sign(hash)
	if err != nil {
		return rawSig, err
	}

	return append(rawSig, pubData...), nil
}

// VerifySignature checks that the given public key created signature over hash.
// The public key should be in compressed (33 bytes) or uncompressed (65 bytes) format.
// The signature should have the 64 byte [R || S] format.
func VerifySignature(pubkey, hash, signature []byte) bool {
	byteLen := (sm2.P256Sm2().Params().BitSize + 7) >> 3
	x, y := elliptic.Unmarshal(S256(), pubkey)
	sm2_pub := &sm2.PublicKey{Curve: S256(), X: x, Y: y}

	return sm2_pub.Verify(hash, signature[:len(signature)-1-2*byteLen])
}

// DecompressPubkey parses a public key in the 33-byte compressed format.
func DecompressPubkey(pubkey []byte) (*ecdsa.PublicKey, error) {
	if len(pubkey) != 33 {
		return nil, fmt.Errorf("compresseed public key length error", len(pubkey))
	}
	x := sm2.Decompress(pubkey).X
	if x == nil {
		return nil, fmt.Errorf("invalid public key")
	}

	y := sm2.Decompress(pubkey).Y
	pk := &ecdsa.PublicKey{Curve:S256(), X: x, Y: y}
	return pk, nil
}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func CompressPubkey(pubkey *ecdsa.PublicKey) []byte {
	pk := &sm2.PublicKey{Curve:S256(), X: pubkey.X, Y: pubkey.Y}
	return sm2.Compress(pk)
}

// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return sm2.P256Sm2()
}

// ValidateSignatureValues verifies whether the signature values are valid with
// the given chain rules. The v value is assumed to be 1.
func ValidateSignatureValues(v byte, r, s *big.Int, homestead bool) bool {
	if r.Cmp(common.Big1) < 0 || s.Cmp(common.Big1) < 0 {
		return false
	}
	return r.Cmp(sm2N) < 0 && s.Cmp(sm2N) < 0
}
