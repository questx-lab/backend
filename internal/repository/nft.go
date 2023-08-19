package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NftRepository interface {
	// NFT
	Create(context.Context, *entity.NonFungibleToken) error
	GetByID(context.Context, int64) (*entity.NonFungibleToken, error)
	GetByIDs(context.Context, []int64) ([]entity.NonFungibleToken, error)
	GetByCommunityID(ctx context.Context, communityID string) ([]entity.NonFungibleToken, error)
	GetByUserID(ctx context.Context, userID string) ([]entity.ClaimedNonFungibleToken, error)
	IncreaseClaimed(ctx context.Context, tokenID int64, amount int) error
	IncreaseTotalBalance(ctx context.Context, tokenID int64, amount int) error

	// History
	CreateHistory(context.Context, *entity.NonFungibleTokenMintHistory) error

	// Claimed
	UpsertClaimedToken(context.Context, *entity.ClaimedNonFungibleToken) error
}

type nftRepository struct {
}

func NewNftRepository() *nftRepository {
	return &nftRepository{}
}

func (r *nftRepository) Create(ctx context.Context, nft *entity.NonFungibleToken) error {
	return xcontext.DB(ctx).Create(nft).Error
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

func (r *nftRepository) IncreaseClaimed(ctx context.Context, tokenID int64, amount int) error {
	tx := xcontext.DB(ctx).Model(&entity.NonFungibleToken{}).
		Where("id=? AND number_of_claimed <= total_balance-?", tokenID, amount).
		Update("number_of_claimed", gorm.Expr("number_of_claimed+?", amount))
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return tx.Error
}

func (r *nftRepository) IncreaseTotalBalance(ctx context.Context, tokenID int64, amount int) error {
	return xcontext.DB(ctx).Model(&entity.NonFungibleToken{}).
		Where("id=?", tokenID).
		Update("total_balance", gorm.Expr("total_balance+?", amount)).Error
}

func (r *nftRepository) GetByIDs(ctx context.Context, ids []int64) ([]entity.NonFungibleToken, error) {
	var result []entity.NonFungibleToken
	if err := xcontext.DB(ctx).Where("id IN (?)", ids).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *nftRepository) GetByUserID(ctx context.Context, userID string) ([]entity.ClaimedNonFungibleToken, error) {
	var result []entity.ClaimedNonFungibleToken
	if err := xcontext.DB(ctx).
		Where("user_id = ?", userID).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *nftRepository) UpsertClaimedToken(ctx context.Context, data *entity.ClaimedNonFungibleToken) error {
	return xcontext.DB(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "non_fungible_token_id"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"amount": gorm.Expr("amount+?", data.Amount),
			}),
		}).Create(data).Error
}
