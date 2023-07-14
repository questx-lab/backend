package ethutil

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

var (
	ethChains = map[string]bool{
		"eth":              true,
		"ropsten-testnet":  true,
		"goerli-testnet":   true,
		"binance-testnet":  true,
		"ganache1":         true,
		"ganache2":         true,
		"fantom-testnet":   true,
		"polygon-testnet":  true,
		"xdai":             true,
		"arbitrum-testnet": true,
		"avaxc-testnet":    true,
	}
	ethSigners map[string]etypes.Signer
)

func init() {
	ethSigners = make(map[string]etypes.Signer)

	ethSigners["eth"] = etypes.NewLondonSigner(GetChainIntFromId("eth"))
	ethSigners["ropsten-testnet"] = etypes.NewLondonSigner(GetChainIntFromId("ropsten-testnet"))
	ethSigners["goerli-testnet"] = etypes.NewLondonSigner(GetChainIntFromId("goerli-testnet"))
	ethSigners["binance-testnet"] = etypes.NewLondonSigner(GetChainIntFromId("binance-testnet"))
	ethSigners["ganache1"] = etypes.NewLondonSigner(GetChainIntFromId("ganache1"))
	ethSigners["ganache2"] = etypes.NewLondonSigner(GetChainIntFromId("ganache2"))
	ethSigners["fantom-testnet"] = etypes.NewLondonSigner(GetChainIntFromId("fantom-testnet"))
	ethSigners["xdai"] = etypes.NewLondonSigner(GetChainIntFromId("xdai"))
	ethSigners["polygon-testnet"] = etypes.NewLondonSigner(GetChainIntFromId("polygon-testnet"))
	ethSigners["arbitrum-testnet"] = etypes.NewLondonSigner(GetChainIntFromId("arbitrum-testnet"))
	ethSigners["avaxc-testnet"] = etypes.NewLondonSigner(GetChainIntFromId("avaxc-testnet"))
}

func GetChainIntFromId(chain string) *big.Int {
	switch chain {
	case "eth":
		return big.NewInt(1)
	case "ropsten-testnet":
		return big.NewInt(3)
	case "goerli-testnet":
		return big.NewInt(5)
	case "binance-testnet":
		return big.NewInt(97)
	case "xdai":
		return big.NewInt(100)
	case "ganache1":
		return big.NewInt(189985)
	case "ganache2":
		return big.NewInt(189986)
	case "fantom-testnet":
		return big.NewInt(4002)
	case "polygon-testnet":
		return big.NewInt(80001)
	case "arbitrum-testnet":
		return big.NewInt(421611)
	case "avaxc-testnet":
		return big.NewInt(43113)

	// Non-evm
	case "cardano-testnet":
		return big.NewInt(98723843487)
	case "solana-devnet":
		return big.NewInt(2382734923)
	default:
		log.Println("unknown chain:", chain)
		return nil
	}
}

func IsETHBasedChain(chain string) bool {
	switch chain {
	case "eth", "ropsten-testnet", "goerli-testnet", "binance-testnet", "ganache1", "ganache2",
		"fantom-testnet", "polygon-testnet", "xdai", "arbitrum-testnet", "avaxc-testnet":
		return true
	}

	return false
}

func GetEthChainSigner(chain string) etypes.Signer {
	return ethSigners[chain]
}

func GetSupportedEthChains() map[string]bool {
	newMap := make(map[string]bool)
	for k, v := range ethChains {
		newMap[k] = v
	}

	return newMap
}

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
	return ecdsa.GenerateKey(crypto.S256(), reader)
}
