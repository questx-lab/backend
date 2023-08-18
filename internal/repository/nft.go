package repository

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NftRepository interface {
	// NFT
	Upsert(context.Context, *entity.NonFungibleToken) error
	GetByID(context.Context, int64) (*entity.NonFungibleToken, error)
	GetByIDs(context.Context, []int64) ([]entity.NonFungibleToken, error)
	GetByCommunityID(ctx context.Context, communityID string) ([]entity.NonFungibleToken, error)
	IncreaseClaimed(ctx context.Context, tokenID, amount, totalBalance int64) error

	// History
	CreateHistory(context.Context, *entity.NonFungibleTokenMintHistory) error
	BalanceOf(context.Context, int64) (int, error)
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

func (r *nftRepository) GetByCommunityID(ctx context.Context, communityID string) ([]entity.NonFungibleToken, error) {
	var result []entity.NonFungibleToken
	if err := xcontext.DB(ctx).Model(&entity.NonFungibleToken{}).
		Where("community_id = ?", communityID).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *nftRepository) CreateHistory(ctx context.Context, e *entity.NonFungibleTokenMintHistory) error {
	return xcontext.DB(ctx).Create(e).Error
}

func (r *nftRepository) BalanceOf(ctx context.Context, id int64) (int, error) {
	var result int64
	err := xcontext.DB(ctx).Model(&entity.NonFungibleTokenMintHistory{}).
		Select("SUM(non_fungible_token_mint_histories.amount)").
		Joins("join blockchain_transactions on blockchain_transactions.id=non_fungible_token_mint_histories.transaction_id").
		Where("non_fungible_token_mint_histories.non_fungible_token_id=?", id).
		Where("blockchain_transactions.status=?", entity.BlockchainTransactionStatusTypeSuccess).
		Group("non_fungible_token_mint_histories.non_fungible_token_id").
		Take(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}

		return 0, err
	}

	return int(result), nil
}

func (r *nftRepository) IncreaseClaimed(ctx context.Context, tokenID, amount, totalBalance int64) error {
	tx := xcontext.DB(ctx).Model(&entity.NonFungibleToken{}).
		Where("id=? AND number_of_claimed <= ?", tokenID, totalBalance-amount).
		Update("number_of_claimed", gorm.Expr("number_of_claimed+?", amount))
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return tx.Error
}

func (r *nftRepository) GetByIDs(ctx context.Context, ids []int64) ([]entity.NonFungibleToken, error) {
	var result []entity.NonFungibleToken
	if err := xcontext.DB(ctx).Model(&entity.NonFungibleToken{}).
		Where("id IN (?)", ids).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}
