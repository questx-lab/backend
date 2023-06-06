package eth

import (
	"context"
	"math/big"

	"github.com/questx-lab/backend/pkg/xcontext"
)

func GetChainIntFromId(ctx context.Context, chain string) *big.Int {
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
	case "fantom-testnet":
		return big.NewInt(4002)
	case "polygon-testnet":
		return big.NewInt(80001)
	case "arbitrum-testnet":
		return big.NewInt(421611)
	case "avaxc-testnet":
		return big.NewInt(43113)

	default:
		xcontext.Logger(ctx).Errorf("unknown chain: %s", chain)
		return nil
	}
}
