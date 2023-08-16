package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type NftSetRepository interface {
	Create(context.Context, *entity.NFTSet) error
	GetByID(context.Context, string) (*entity.NFTSet, error)
}

type nftSetRepository struct {
}

func NewNftSetRepository() *nftSetRepository {
	return &nftSetRepository{}
}

func (r *nftSetRepository) Create(ctx context.Context, set *entity.NFTSet) error {
	return xcontext.DB(ctx).Create(set).Error
}

func (r *nftSetRepository) GetByID(ctx context.Context, id string) (*entity.NFTSet, error) {
	var result entity.NFTSet
	err := xcontext.DB(ctx).Where("id = ?", id).Take(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}
