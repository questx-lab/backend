package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/ethutil"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type signer struct {
	client EthClient
	cfg    config.EthConfigs
}

func (s *signer) CreateTransaction(ctx context.Context, from, to, chain string, amount *big.Int) (*types.Transaction, error) {
	fromAddress := common.HexToAddress(to)
	toAddress := common.HexToAddress(to)

	gasLimit := uint64(21000) // in units
	gasPrice, err := s.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	nonce, err := s.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, err
	}

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, data)

	privateKey, err := crypto.HexToECDSA(s.cfg.Keys.PrivKey)
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, getSignerByChainID(ctx, chain), privateKey)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func getSignerByChainID(ctx context.Context, chain string) types.Signer {
	switch chain {
	case "eth":
		return types.NewLondonSigner(ethutil.GetChainIntFromId(chain))
	case "ropsten-testnet":
		return types.NewEIP155Signer(ethutil.GetChainIntFromId(chain))
	case "goerli-testnet":
		return types.NewLondonSigner(ethutil.GetChainIntFromId(chain))
	case "binance-testnet":
		return types.NewLondonSigner(ethutil.GetChainIntFromId(chain))
	case "xdai":
		return types.NewLondonSigner(ethutil.GetChainIntFromId(chain))
	case "fantom-testnet":
		return types.NewLondonSigner(ethutil.GetChainIntFromId(chain))
	case "polygon-testnet":
		return types.NewLondonSigner(ethutil.GetChainIntFromId(chain))
	case "arbitrum-testnet":
		return types.NewLondonSigner(ethutil.GetChainIntFromId(chain))
	case "avaxc-testnet":
		return types.NewLondonSigner(ethutil.GetChainIntFromId(chain))

	default:
		xcontext.Logger(ctx).Errorf("unknown chain: %s", chain)
		return nil
	}
}
