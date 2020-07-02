// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package poloclient

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/HiNounou029/nounouchain/crypto/sha3"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/mattn/go-tty"
	"github.com/pkg/errors"
	"hash"
	"io/ioutil"
	"strings"
)

type BlockRef [8]byte

type Bytes32 [32]byte

type Address [20]byte

func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

func FromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

func NewBlake2b() hash.Hash {
	hash := sha3.NewKeccak256()
	return hash
}

func (b Bytes32) Bytes() []byte {
	return b[:]
}

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

func Bytes2Hex(d []byte) string {
	return hex.EncodeToString(d)
}

func (a Address) String() string {
	return "0x" + hex.EncodeToString(a[:])
}

func BytesToBytes32(b []byte) Bytes32 {
	return Bytes32(BytesToHash(b))
}

// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [HashLength]byte

// BytesToHash sets b to hash.
// If b is larger than len(h), b will be cropped from the left.
func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

func ParseAddress(s string) (Address, error) {
	if len(s) == AddressLength*2 {
	} else if len(s) == AddressLength*2+2 {
		if strings.ToLower(s[:2]) != "0x" {
			return Address{}, errors.New("invalid prefix")
		}
		s = s[2:]
	} else {
		return Address{}, errors.New("invalid length")
	}

	var addr Address
	_, err := hex.Decode(addr[:], []byte(s))
	if err != nil {
		return Address{}, err
	}
	return addr, nil
}

func MustParseAddress(s string) Address {
	addr, err := ParseAddress(s)
	if err != nil {
		panic(err)
	}
	return addr
}

func readPasswordFromNewTTY(prompt string) (string, error) {
	t, err := tty.Open()
	if err != nil {
		return "", err
	}
	defer t.Close()
	fmt.Fprint(t.Output(), prompt)
	pass, err := t.ReadPasswordNoEcho()
	if err != nil {
		return "", err
	}
	return pass, err
}

func loadPrivateKeyInKeyStore(path string) (*ecdsa.PrivateKey, error) {
	keyjson, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(keyjson, &map[string]interface{}{}); err != nil {
		return nil, errors.WithMessage(err, "unmarshal")
	}
	password, err := readPasswordFromNewTTY("Enter loading passphrase: ")
	if err != nil {
		return nil, err
	}

	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		return nil, errors.WithMessage(err, "decrypt")
	}

	fmt.Printf("path[%s] keystore imported\n", path)
	return key.PrivateKey, nil
}

func MustLoadPrivateKeyStore(path string) *ecdsa.PrivateKey {
	if len(path) <= 0 {
		panic("empty path")
	}
	key, err := loadPrivateKeyInKeyStore(path)
	if err != nil {
		panic(err)
	}
	return key
}
