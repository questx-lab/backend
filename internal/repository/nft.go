package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm/clause"
)

type NftRepository interface {
	Upsert(context.Context, *entity.NonFungibleToken) error
	GetByID(context.Context, int64) (*entity.NonFungibleToken, error)
}

type nftRepository struct {
}

func NewNftRepository() *nftRepository {
	return &nftRepository{}
}

func (r *nftRepository) Upsert(ctx context.Context, nft *entity.NonFungibleToken) error {
	return xcontext.DB(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "id"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"title":       nft.Title,
				"description": nft.Description,
				"image_url":   nft.ImageUrl,
			}),
		}).Create(nft).Error
}

func (r *nftRepository) GetByID(ctx context.Context, id int64) (*entity.NonFungibleToken, error) {
	var result entity.NonFungibleToken
	err := xcontext.DB(ctx).Where("id = ?", id).Take(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}
