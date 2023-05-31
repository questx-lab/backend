package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type BlockChainTransactionRepository interface {
	CreateTransaction(ctx context.Context, e *entity.BlockChainTransaction) error
}

type blockChainTransactionRepository struct {
}

func (r *blockChainTransactionRepository) CreateTransaction(ctx context.Context, e *entity.BlockChainTransaction) error {
	if err := xcontext.DB(ctx).Model(e).Create(e).Error; err != nil {
		return err
	}
	return nil
}
