package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type NftMintHistoryRepository interface {
	Create(context.Context, *entity.NonFungibleTokenMintHistory) error
	GetAllPending(ctx context.Context) ([]entity.NonFungibleTokenMintHistory, error)
	AggregateByNftID(ctx context.Context, nftID int64) (*AggregateNFT, error)
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

type AggregateNFT struct {
	PendingAmount int64
	ActiveAmount  int64
	FailureAmount int64
}

func (r *nftMintHistoryRepository) AggregateByNftID(ctx context.Context, nftID int64) (*AggregateNFT, error) {
	var aggResult []struct {
		Status string
		Count  int64
	}

	if err := xcontext.DB(ctx).Model(&entity.NonFungibleTokenMintHistory{}).
		Select("blockchain_transactions.status", "COUNT(blockchain_transactions.status) as count").
		Joins("JOIN blockchain_transactions ON blockchain_transactions.id = non_fungible_token_mint_histories.transaction_id").
		Where("non_fungible_token_id = ?", nftID).
		Group("blockchain_transactions.status").Find(&aggResult).Error; err != nil {
		return nil, err
	}

	var result AggregateNFT
	for _, val := range aggResult {
		switch entity.BlockchainTransactionStatusType(val.Status) {
		case entity.BlockchainTransactionStatusTypeSuccess:
			result.ActiveAmount = val.Count
		case entity.BlockchainTransactionStatusTypeInProgress:
			result.PendingAmount = val.Count
		case entity.BlockchainTransactionStatusTypeFailure:
			result.FailureAmount = val.Count
		}
	}

	return &result, nil
}
