package aescrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	crand "crypto/rand"
	"golang.org/x/crypto/scrypt"
)

func ensureInt(x interface{}) int {
	res, ok := x.(int)
	if !ok {
		res = int(x.(float64))
	}
	return res
}

func GetEntropyCSPRNG(n int) []byte {
	mainBuff := make([]byte, n)
	_, err := io.ReadFull(crand.Reader, mainBuff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return mainBuff
}

func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	// AES-128 is selected due to size of encryptKey.
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}

func EncryptString(strData string, auth string) ([]byte, error) {
	authArray := []byte(auth)
	salt := GetEntropyCSPRNG(32)
	derivedKey, err := scrypt.Key(authArray, salt, 1<<18, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	encryptKey := derivedKey[:16]

	iv := GetEntropyCSPRNG(aes.BlockSize) // 16
	cipherText, err := aesCTRXOR(encryptKey, []byte(strData), iv)
	if err != nil {
		return nil, err
	}

	s256 := sha256.New()	//crypto.Keccak256(derivedKey[16:32], cipherText)
	s256.Write(derivedKey[16:32])
	s256.Write(cipherText)
	hash := s256.Sum(nil)

	scryptParamsJSON := make(map[string]interface{}, 5)
	scryptParamsJSON["n"] = 1<<18
	scryptParamsJSON["r"] = 8
	scryptParamsJSON["p"] = 1
	scryptParamsJSON["dklen"] = 32
	scryptParamsJSON["salt"] = hex.EncodeToString(salt)

	cipherParamsJSON := make(map[string]interface{},1)
	cipherParamsJSON["IV"] = hex.EncodeToString(iv)

	cryptoStruct := make(map[string]interface{}, 6)
	cryptoStruct["Cipher"] = "aes-128-ctr"
	cryptoStruct["CipherText"] = hex.EncodeToString(cipherText)
	cryptoStruct["CipherParams"] = cipherParamsJSON
	cryptoStruct["KDF"] = "aescrypto"
	cryptoStruct["KDFParams"] = scryptParamsJSON
	cryptoStruct["MAC"] = hex.EncodeToString(hash)

	return json.Marshal(cryptoStruct)
}

func DecryptCryptoJSON(cryptoJson []byte, auth string) (dataBytes []byte, err error) {
	var data map[string]interface{}
	json.Unmarshal(cryptoJson, &data)

	if data["Cipher"] != "aes-128-ctr" {
		return nil, fmt.Errorf("Cipher not supported: %v", data["Cipher"])
	}

	hash, err := hex.DecodeString(data["MAC"].(string))
	if err != nil {
		return nil, err
	}

	cipherText, err := hex.DecodeString(data["CipherText"].(string))
	if err != nil {
		return nil, err
	}

	//
	var iv []byte
	if v, ok := data["CipherParams"]; ok {
		cipherParams := v.(map[string]interface{})
		iv, err = hex.DecodeString(cipherParams["IV"].(string))
		if err != nil {
			return nil, err
		}
	}

	derivedKey, err := getKDFKey(data, auth)
	if err != nil {
		return nil, err
	}

	s256 := sha256.New()	//crypto.Keccak256(derivedKey[16:32], cipherText)
	s256.Write(derivedKey[16:32])
	s256.Write(cipherText)
	calcHash := s256.Sum(nil)
	if !bytes.Equal(calcHash, hash) {
		return nil, errors.New("could not decrypt key with given passphrase")
	}

	plainText, err := aesCTRXOR(derivedKey[:16], cipherText, iv)
	if err != nil {
		return nil, err
	}
	return plainText, err
}

func getKDFKey(cryptoJson map[string]interface{}, auth string) ([]byte, error) {
	authArray := []byte(auth)

	if v, ok := cryptoJson["KDFParams"]; ok {
		KDFParams := v.(map[string]interface{})

		salt, err := hex.DecodeString(KDFParams["salt"].(string))
		if err != nil {
			return nil, err
		}

		dkLen := ensureInt(KDFParams["dklen"])

		if cryptoJson["KDF"] == "aescrypto" {
			n := ensureInt(KDFParams["n"])
			r := ensureInt(KDFParams["r"])
			p := ensureInt(KDFParams["p"])
			return scrypt.Key(authArray, salt, n, r, p, dkLen)
		}
		//else if cryptoJSON.KDF == "pbkdf2" {
		//	c := ensureInt(cryptoJSON.KDFParams["c"])
		//	prf := cryptoJSON.KDFParams["prf"].(string)
		//	if prf != "hmac-sha256" {
		//		return nil, fmt.Errorf("Unsupported PBKDF2 PRF: %s", prf)
		//	}
		//	key := pbkdf2.Key(authArray, salt, c, dkLen, sha256.New)
		//	return key, nil
		//}
	}

	return nil, fmt.Errorf("Unsupported KDF: %s", cryptoJson["KDF"])
}