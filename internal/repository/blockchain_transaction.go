package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type BlockChainTransactionRepository interface {
	CreateTransaction(ctx context.Context, e *entity.BlockchainTransaction) error
	UpdateByTxHash(ctx context.Context, txHash, chain string, data *entity.BlockchainTransaction) error
	GetByTxHash(ctx context.Context, txHash, chain string) (*entity.BlockchainTransaction, error)
}

type blockChainTransactionRepository struct {
}

func NewBlockChainTransactionRepository() *blockChainTransactionRepository {
	return &blockChainTransactionRepository{}
}

func (r *blockChainTransactionRepository) CreateTransaction(ctx context.Context, e *entity.BlockchainTransaction) error {
	if err := xcontext.DB(ctx).Model(e).Create(e).Error; err != nil {
		return err
	}
	return nil
}

func (r *blockChainTransactionRepository) UpdateByTxHash(ctx context.Context, txHash, chain string, data *entity.BlockchainTransaction) error {
	return xcontext.DB(ctx).Model(&entity.BlockchainTransaction{}).Where("tx_hash = ? AND chain = ?", txHash, chain).Updates(data).Error
}

func (r *blockChainTransactionRepository) GetByTxHash(ctx context.Context, txHash, chain string) (*entity.BlockchainTransaction, error) {
	var result entity.BlockchainTransaction
	if err := xcontext.DB(ctx).Take(&result, "tx_hash = ? AND chain = ?", txHash, chain).Error; err != nil {
		return nil, err
	}

	return &result, nil
}
