package ethutil

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

func PublicKeyBytesToAddress(publicKey []byte) common.Address {
	var buf []byte

	hash := sha3.NewLegacyKeccak256()
	hash.Write(publicKey[1:]) // remove EC prefix 04
	buf = hash.Sum(nil)
	address := buf[12:]

	return common.HexToAddress(hex.EncodeToString(address))
}

func GeneratePrivateKey(secret, nonce []byte) (*ecdsa.PrivateKey, error) {
	seed := sha256.Sum256(append(secret, nonce...))
	log.Println("secret", string(secret))
	log.Println("nonce", string(nonce))
	log.Println("seed", string(seed[:]))
	randomSeed := bytes.Repeat(seed[:], 2)
	log.Println("randomSeed", string(randomSeed))
	reader := bytes.NewReader(randomSeed)
	key, err := ecdsa.GenerateKey(crypto.S256(), reader)
	log.Println("key", key.PublicKey)
	return key, err
}
