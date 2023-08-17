package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type NftMintHistoryRepository interface {
	Create(context.Context, *entity.NonFungibleTokenMintHistory) error
	GetAllPending(ctx context.Context) ([]entity.NonFungibleTokenMintHistory, error)
}

type nftMintHistoryRepository struct{}

func NewNftMintHistoryRepository() *nftMintHistoryRepository {
	return &nftMintHistoryRepository{}
}

func (r *nftMintHistoryRepository) Create(ctx context.Context, e *entity.NonFungibleTokenMintHistory) error {
	return xcontext.DB(ctx).Create(e).Error
}

func (r *nftMintHistoryRepository) GetAllPending(ctx context.Context) ([]entity.NonFungibleTokenMintHistory, error) {
	var result []entity.NonFungibleTokenMintHistory
	err := xcontext.DB(ctx).Model(&entity.NonFungibleTokenMintHistory{}).
		Joins("JOIN nft_sets ON non_fungible_tokens.id = non_fungible_token_mint_histories.non_fungible_token_id").
		Where("transaction_id IS NULL").
		Find(&result).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}
