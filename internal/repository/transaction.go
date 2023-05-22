package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/blockchain/types"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type TransactionRepository interface {
	Create(context.Context, *entity.Transaction) error
	GetByID(context.Context, string) (*entity.Transaction, error)
	GetByUserID(context.Context, string) ([]entity.Transaction, error)

	/////// for blockchain ///////
	SaveTxs(chain string, blockHeight int64, txs *types.Txs)
	// Vault address
	SetVault(chain, address string, token string) error
	GetVaults(chain string) ([]string, error)
}

type transactionRepository struct{}

func NewTransactionRepository() *transactionRepository {
	return &transactionRepository{}
}

func (r *transactionRepository) Create(ctx context.Context, tx *entity.Transaction) error {
	return xcontext.DB(ctx).Create(tx).Error
}

func (r *transactionRepository) GetByID(ctx context.Context, id string) (*entity.Transaction, error) {
	var result entity.Transaction
	if err := xcontext.DB(ctx).Take(&result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *transactionRepository) GetByUserID(ctx context.Context, userID string) ([]entity.Transaction, error) {
	var result []entity.Transaction
	if err := xcontext.DB(ctx).Find(&result, "user_id=?", userID).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *transactionRepository) SaveTxs(chain string, blockHeight int64, txs *types.Txs) {
	panic("unimplemented")

}

func (r *transactionRepository) SetVault(chain, address string, token string) error {
	panic("unimplemented")
}

func (r *transactionRepository) GetVaults(chain string) ([]string, error) {
	panic("unimplemented")
}
