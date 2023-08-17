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
	AggregateByCommunityID(ctx context.Context, communityID string) ([]*AggregateCommunityNFT, error)
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

type AggregateCommunityNFT struct {
	NftID       int64
	Title       string
	Description string
	ImageUrl    string

	PendingAmount int64
	ActiveAmount  int64
	FailureAmount int64
}

func (r *nftMintHistoryRepository) AggregateByCommunityID(ctx context.Context, communityID string) ([]*AggregateCommunityNFT, error) {
	var aggResult []struct {
		NftID       int64
		Title       string
		Description string
		ImageUrl    string

		Status string
		Count  int64
	}

	if err := xcontext.DB(ctx).Model(&entity.NonFungibleTokenMintHistory{}).
		Select(`
		  non_fungible_tokens.id,
			non_fungible_tokens.title,
			non_fungible_tokens.description,
			non_fungible_tokens.image_url,
		  blockchain_transactions.status,
		  COUNT(*) as count`).
		Joins("JOIN blockchain_transactions ON blockchain_transactions.id = non_fungible_token_mint_histories.transaction_id").
		Joins("JOIN non_fungible_tokens ON non_fungible_tokens.id = non_fungible_token_mint_histories.non_fungible_token_id").
		Where("non_fungible_tokens.community_id = ?", communityID).
		Group("non_fungible_tokens.id, blockchain_transactions.status").Find(&aggResult).Error; err != nil {
		return nil, err
	}

	nftMap := make(map[int64]*AggregateCommunityNFT)

	for _, val := range aggResult {
		if _, ok := nftMap[val.NftID]; !ok {
			nftMap[val.NftID] = &AggregateCommunityNFT{
				Title:       val.Title,
				Description: val.Description,
				ImageUrl:    val.ImageUrl,
				NftID:       val.NftID,
			}

		}
		switch entity.BlockchainTransactionStatusType(val.Status) {
		case entity.BlockchainTransactionStatusTypeSuccess:
			nftMap[val.NftID].ActiveAmount = val.Count
		case entity.BlockchainTransactionStatusTypeInProgress:
			nftMap[val.NftID].PendingAmount = val.Count
		case entity.BlockchainTransactionStatusTypeFailure:
			nftMap[val.NftID].FailureAmount = val.Count
		}
	}

	var result []*AggregateCommunityNFT

	for _, val := range nftMap {
		tempt := *val
		result = append(result, &tempt)
	}

	return result, nil
}
