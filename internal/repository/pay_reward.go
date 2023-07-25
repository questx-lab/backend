package repository

import (
	"context"
	"database/sql"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type PayRewardRepository interface {
	Create(context.Context, *entity.PayReward) error
	GetByID(context.Context, string) (*entity.PayReward, error)
	GetByUserID(context.Context, string) ([]entity.PayReward, error)
	UpdateTransactionByID(ctx context.Context, id string, transactionID sql.NullString) error
	UpdateTransactionByIDs(ctx context.Context, ids []string, transactionID sql.NullString) error
	GetAllPending(context.Context) ([]entity.PayReward, error)
}

type payRewardRepository struct{}

func NewPayRewardRepository() *payRewardRepository {
	return &payRewardRepository{}
}

func (r *payRewardRepository) Create(ctx context.Context, tx *entity.PayReward) error {
	return xcontext.DB(ctx).Create(tx).Error
}

func (r *payRewardRepository) GetByID(ctx context.Context, id string) (*entity.PayReward, error) {
	var result entity.PayReward
	if err := xcontext.DB(ctx).Take(&result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *payRewardRepository) GetByUserID(ctx context.Context, userID string) ([]entity.PayReward, error) {
	var result []entity.PayReward
	if err := xcontext.DB(ctx).Find(&result, "to_user_id=?", userID).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *payRewardRepository) UpdateTransactionByID(ctx context.Context, id string, transactionID sql.NullString) error {
	return xcontext.DB(ctx).
		Model(&entity.PayReward{}).
		Where("id = ?", id).
		Update("transaction_id", transactionID).Error
}

func (r *payRewardRepository) UpdateTransactionByIDs(ctx context.Context, ids []string, transactionID sql.NullString) error {
	return xcontext.DB(ctx).
		Model(&entity.PayReward{}).
		Where("id IN (?)", ids).
		Update("transaction_id", transactionID).Error
}

func (r *payRewardRepository) GetAllPending(ctx context.Context) ([]entity.PayReward, error) {
	var result []entity.PayReward
	err := xcontext.DB(ctx).Model(&entity.PayReward{}).
		Where("transaction_id IS NULL").
		Find(&result).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}
