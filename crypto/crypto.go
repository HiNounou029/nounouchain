// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/common/rlp"
	"github.com/go-hmac-drbg/hmacdrbg"
	"golang.org/x/crypto/sha3"
)

var (
	secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	sm2N, _        = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFF7203DF6B21C6052B53BBF40939D54123", 16)
	secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
	one            = new(big.Int).SetInt64(1)
)

var errInvalidPubkey = errors.New("invalid secp256k1 public key")

// Keccak512 calculates and returns the SHA3-512 hash of the input data.
func Keccak512(data ...[]byte) []byte {
	d := sha3.New512()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// CreateAddress creates an address given the bytes and the nonce
func CreateAddress(b common.Address, nonce uint64) common.Address {
	data, _ := rlp.EncodeToBytes([]interface{}{b, nonce})
	return common.BytesToAddress(Keccak256(data)[12:])
}

// CreateAddress2 creates an address given the address bytes, initial
// contract code hash and a salt.
func CreateAddress2(b common.Address, salt [32]byte, inithash []byte) common.Address {
	return common.BytesToAddress(Keccak256([]byte{0xff}, b.Bytes(), salt[:], inithash)[12:])
}

// ToECDSA creates a private key with the given D value.
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	return toECDSA(d, true)
}

// ToECDSAUnsafe blindly converts a binary blob to a private key. It should almost
// never be used unless you are sure the input is valid and want to avoid hitting
// errors due to bad origin encoding (0 prefixes cut off).
func ToECDSAUnsafe(d []byte) *ecdsa.PrivateKey {
	priv, _ := toECDSA(d, false)
	return priv
}

// toECDSA creates a private key with the given D value. The strict parameter
// controls whether the key's length should be enforced at the curve size or
// it can also accept legacy encodings (0 prefixes).
func toECDSA(d []byte, strict bool) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if priv.D.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}

// FromECDSA exports a private key into a binary dump.
func FromECDSA(priv *ecdsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return math.PaddedBigBytes(priv.D, priv.Params().BitSize/8)
}

// UnmarshalPubkey converts bytes to a secp256k1/sm2 public key.
func UnmarshalPubkey(pub []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(S256(), pub)
	if x == nil {
		return nil, errInvalidPubkey
	}
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y)
}

// HexToECDSA parses a secp256k1/sm2 private key.
func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	return ToECDSA(b)
}

// LoadECDSA loads a secp256k1/sm2 private key from the given file.
func LoadECDSA(file string) (*ecdsa.PrivateKey, error) {
	buf := make([]byte, 64)
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	if _, err := io.ReadFull(fd, buf); err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}
	return ToECDSA(key)
}

// SaveECDSA saves a secp256k1/sm2 private key to the given file with
// restrictive permissions. The key data is saved hex-encoded.
func SaveECDSA(file string, key *ecdsa.PrivateKey) error {
	k := hex.EncodeToString(FromECDSA(key))
	return ioutil.WriteFile(file, []byte(k), 0600)
}

func PubkeyToAddress(p ecdsa.PublicKey) common.Address {
	pubBytes := FromECDSAPub(&p)
	return common.BytesToAddress(Keccak256(pubBytes[1:])[12:])
}

func zeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}

func GenerateKey() (*ecdsa.PrivateKey, error) {
	params := S256().Params()
	securityLevel := params.BitSize / 2
	outlen := params.BitSize / 8
	k := hmac_drbc(securityLevel, outlen)
	n := new(big.Int).Sub(params.N, one)
	k.Mod(k, n)
	k.Add(k, one)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	priv.D = k
	priv.PublicKey.X, priv.PublicKey.Y = S256().ScalarBaseMult(k.Bytes())
	return priv, nil
}

// hmac_drbc generate a random element of the field underlying the given
// curve using the procedure given in
// "Suit B implementers’ guide to FIPS 186-3(ecdsa) A2.1."
// Refer to https://github.com/cruxic/go-hmac-drbg
func hmac_drbc(securityLevel int, outlen int) (r *big.Int) {
	seed := make([]byte, outlen)
	_, err := rand.Read(seed)
	if err != nil {
		errors.New("rand(seed) error")
	}
	//Note: security-level must be one of: 112, 128, 192, 256.
	//Initial seed length must be at least 1.5 times security-level
	// (max is hmacdrbg.MaxEntropyBytes)
	rng := hmacdrbg.NewHmacDrbg(securityLevel, seed, nil)
	randData := make([]byte, outlen)
	for i := 0; i < 1; i++ {
		//Note: Generate cannot do more than
		// hmacdrbg.MaxBytesPerGenerate (937 bytes)
		if !rng.Generate(randData) {
			//Reseed is required every 10,000 calls to Generate (sooner is fine)
			_, err = rand.Read(seed)
			if err != nil {
				errors.New("rand(seed) error")
			}
			err = rng.Reseed(seed)
			if err != nil {
				errors.New("Reseed(seed) error")
			}
		}
		r := new(big.Int).SetBytes(randData)
		return r
	}
	return
}
