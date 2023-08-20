package ethutil

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
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
	randomSeed := bytes.Repeat(seed[:], 2)
	reader := bytes.NewReader(randomSeed)
	return ecdsa.GenerateKey(ethcrypto.S256(), reader)
}

func GeneratePublicKey(secret, nonce []byte) (common.Address, error) {
	walletPrivateKey, err := GeneratePrivateKey(secret, nonce)
	if err != nil {
		return common.Address{}, err
	}

	return crypto.PubkeyToAddress(walletPrivateKey.PublicKey), nil
}
