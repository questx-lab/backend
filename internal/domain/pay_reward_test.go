package domain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"log"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func Test_payRewardDomain_getDispatchedTxRequest(t *testing.T) {
	privateKey, err := crypto.HexToECDSA("7886876e514713dcdc516d1d5a4bde14db8027ae67707303a14d09ba7c409ad4")
	require.NoError(t, err)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	str := hex.EncodeToString(publicKeyBytes)
	log.Println(str)
	////
	publicKeyBytesNew, err := hex.DecodeString(str)
	require.NoError(t, err)
	require.Equal(t, publicKeyBytes, publicKeyBytesNew)
}
