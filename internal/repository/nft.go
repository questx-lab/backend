package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type NftRepository interface {
	Create(context.Context, *entity.NFT) error
	BulkInsert(context.Context, []*entity.NFT) error
	GetByID(context.Context, string) (*entity.NFT, error)
}

type nftRepository struct {
}

func NewNftRepository() *nftRepository {
	return &nftRepository{}
}

func (r *nftRepository) Create(ctx context.Context, nft *entity.NFT) error {
	return xcontext.DB(ctx).Create(nft).Error
}

func (r *nftRepository) BulkInsert(ctx context.Context, nfts []*entity.NFT) error {
	return xcontext.DB(ctx).Create(nfts).Error
}

func (r *nftRepository) GetByID(ctx context.Context, id string) (*entity.NFT, error) {
	var result entity.NFT
	err := xcontext.DB(ctx).Where("id = ?", id).Take(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}
