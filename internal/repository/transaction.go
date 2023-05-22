package repository

import "github.com/questx-lab/backend/pkg/blockchain/types"

type TransactionRepository interface {
	SaveTxs(chain string, blockHeight int64, txs *types.Txs)

	// Vault address
	SetVault(chain, address string, token string) error
	GetVaults(chain string) ([]string, error)
}
