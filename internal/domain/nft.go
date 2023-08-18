package domain

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type NFTDomain interface {
	CreateNFT(context.Context, *model.CreateNFTRequest) (*model.CreateNFTResponse, error)
	GetNFT(context.Context, *model.GetNFTRequest) (*model.GetNFTResponse, error)
	GetNFTs(context.Context, *model.GetNFTsRequest) (*model.GetNFTsResponse, error)
	GetNFTsByMe(context.Context, *model.GetNFTsByMeRequest) (*model.GetNFTsByMeResponse, error)
	GetNFTsByCommunity(context.Context, *model.GetNFTsByCommunityRequest) (*model.GetNFTsByCommunityResponse, error)
}

type nftDomain struct {
	communityRoleVerifier *common.CommunityRoleVerifier
	blockchainCaller      client.BlockchainCaller
	nftRepo               repository.NftRepository
	communityRepo         repository.CommunityRepository
}

func NewNftDomain(
	communityRoleVerifier *common.CommunityRoleVerifier,
	blockchainCaller client.BlockchainCaller,
	nftRepo repository.NftRepository,
	communityRepo repository.CommunityRepository,
) *nftDomain {
	return &nftDomain{
		communityRoleVerifier: communityRoleVerifier,
		blockchainCaller:      blockchainCaller,
		nftRepo:               nftRepo,
		communityRepo:         communityRepo,
	}
}

func (d *nftDomain) CreateNFT(ctx context.Context, req *model.CreateNFTRequest) (*model.CreateNFTResponse, error) {
	var id int64
	var communityID string
	if req.ID == 0 {
		community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found community")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}

		communityID = community.ID
		id = xcontext.SnowFlake(ctx).Generate().Int64()
	} else {
		nft, err := d.nftRepo.GetByID(ctx, req.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found token")
			}

			xcontext.Logger(ctx).Errorf("Unable to create nft set: %v", err)
			return nil, errorx.Unknown
		}

		id = req.ID
		communityID = nft.CommunityID
	}

	if err := d.communityRoleVerifier.Verify(ctx, communityID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	nft := &entity.NonFungibleToken{
		SnowFlakeBase: entity.SnowFlakeBase{ID: id},
		CommunityID:   communityID,
		Title:         req.Title,
		Description:   req.Description,
		ImageUrl:      req.ImageUrl,
		CreatedBy:     xcontext.RequestUserID(ctx),
		Chain:         req.Chain,
	}
	if err := d.nftRepo.Upsert(ctx, nft); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to create nft set: %v", err)
		return nil, errorx.Unknown
	}

	if req.Amount > 0 {
		txID, err := d.blockchainCaller.MintNFT(ctx, communityID, req.Chain, id, req.Amount)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Unable to mint nft: %v", err)
			return nil, errorx.Unknown
		}

		nftMintHistory := &entity.NonFungibleTokenMintHistory{
			NonFungibleTokenID: nft.ID,
			TransactionID:      txID,
			Amount:             int(req.Amount),
		}

		if err := d.nftRepo.CreateHistory(ctx, nftMintHistory); err != nil {
			xcontext.Logger(ctx).Errorf("Unable to create nft mint history: %v", err)
			return nil, errorx.Unknown
		}
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.CreateNFTResponse{}, nil
}

func (d *nftDomain) GetNFT(ctx context.Context, req *model.GetNFTRequest) (*model.GetNFTResponse, error) {
	nft, err := d.nftRepo.GetByID(ctx, req.NftID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found nft")
		}

		xcontext.Logger(ctx).Errorf("Unable to create nft set: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetNFTResponse{NFT: model.ConvertNFT(nft, 0)}, nil
}

func (d *nftDomain) GetNFTsByCommunity(ctx context.Context, req *model.GetNFTsByCommunityRequest) (*model.GetNFTsByCommunityResponse, error) {
	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	nfts, err := d.nftRepo.GetByCommunityID(ctx, community.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get nft by community id: %v", err)
		return nil, errorx.Unknown
	}

	result := []model.NonFungibleToken{}
	for _, nft := range nfts {
		totalBalance, err := d.nftRepo.BalanceOf(ctx, nft.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get total balance of nft: %v", err)
			return nil, errorx.Unknown
		}

		result = append(result, model.ConvertNFT(&nft, totalBalance))
	}

	return &model.GetNFTsByCommunityResponse{NFTs: result}, nil
}

func (d *nftDomain) GetNFTs(ctx context.Context, req *model.GetNFTsRequest) (*model.GetNFTsResponse, error) {
	nfts, err := d.nftRepo.GetByIDs(ctx, req.NftIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found nft")
		}

		xcontext.Logger(ctx).Errorf("Unable to create nft set: %v", err)
		return nil, errorx.Unknown
	}

	result := []model.NonFungibleToken{}
	for _, nft := range nfts {
		totalBalance, err := d.nftRepo.BalanceOf(ctx, nft.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get total balance of nft: %v", err)
			return nil, errorx.Unknown
		}

		result = append(result, model.ConvertNFT(&nft, totalBalance))
	}
	return &model.GetNFTsResponse{NFTs: result}, nil
}

func (d *nftDomain) GetNFTsByMe(ctx context.Context, req *model.GetNFTsByMeRequest) (*model.GetNFTsByMeResponse, error) {
	userID := xcontext.RequestUserID(ctx)
	nfts, err := d.nftRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found nft")
		}

		xcontext.Logger(ctx).Errorf("Unable to create nft set: %v", err)
		return nil, errorx.Unknown
	}

	result := []model.NonFungibleToken{}
	for _, nft := range nfts {
		totalBalance, err := d.nftRepo.BalanceOf(ctx, nft.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get total balance of nft: %v", err)
			return nil, errorx.Unknown
		}

		result = append(result, model.ConvertNFT(&nft, totalBalance))
	}
	return &model.GetNFTsByMeResponse{NFTs: result}, nil
}
