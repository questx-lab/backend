package ethutil

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

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

	return ethcrypto.PubkeyToAddress(walletPrivateKey.PublicKey), nil
}
