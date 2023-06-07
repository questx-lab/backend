package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type PayRewardRepository interface {
	Create(context.Context, *entity.PayReward) error
	GetByID(context.Context, string) (*entity.PayReward, error)
	GetByUserID(context.Context, string) ([]entity.PayReward, error)
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
	if err := xcontext.DB(ctx).Find(&result, "user_id=?", userID).Error; err != nil {
		return nil, err
	}

	return result, nil
}
