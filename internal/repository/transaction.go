package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type TransactionRepository interface {
	Create(xcontext.Context, *entity.Transaction) error
	GetByID(xcontext.Context, string) (*entity.Transaction, error)
	GetByUserID(xcontext.Context, string) ([]entity.Transaction, error)
}

type transactionRepository struct{}

func NewTransactionRepository() *transactionRepository {
	return &transactionRepository{}
}

func (r *transactionRepository) Create(ctx xcontext.Context, tx *entity.Transaction) error {
	return ctx.DB().Create(tx).Error
}

func (r *transactionRepository) GetByID(ctx xcontext.Context, id string) (*entity.Transaction, error) {
	var result entity.Transaction
	if err := ctx.DB().Take(&result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *transactionRepository) GetByUserID(ctx xcontext.Context, userID string) ([]entity.Transaction, error) {
	var result []entity.Transaction
	if err := ctx.DB().Find(&result, "user_id=?", userID).Error; err != nil {
		return nil, err
	}

	return result, nil
}
