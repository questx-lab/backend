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
}

type nftDomain struct {
	communityRoleVerifier *common.CommunityRoleVerifier
	blockchainCaller      client.BlockchainCaller
	nftRepo               repository.NftRepository
	nftMintHistoryRepo    repository.NftMintHistoryRepository
	communityRepo         repository.CommunityRepository
}

func NewNftDomain(
	communityRoleVerifier *common.CommunityRoleVerifier,
	blockchainCaller client.BlockchainCaller,
	nftRepo repository.NftRepository,
	nftMintHistoryRepo repository.NftMintHistoryRepository,
	communityRepo repository.CommunityRepository,
) *nftDomain {
	return &nftDomain{
		communityRoleVerifier: communityRoleVerifier,
		blockchainCaller:      blockchainCaller,
		nftRepo:               nftRepo,
		nftMintHistoryRepo:    nftMintHistoryRepo,
		communityRepo:         communityRepo,
	}
}

func (d *nftDomain) CreateNFT(ctx context.Context, req *model.CreateNFTRequest) (*model.CreateNFTResponse, error) {
	userID := xcontext.RequestUserID(ctx)

	var (
		id          int64
		communityID string
	)

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

	if _, err := d.blockchainCaller.MintNFT(ctx, communityID, req.Chain, id, int(req.Amount)); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to mint nft: %v", err)
		return nil, errorx.Unknown
	}

	nft := &entity.NonFungibleToken{
		SnowFlakeBase: entity.SnowFlakeBase{ID: id},
		CommunityID:   communityID,
		Title:         req.Title,
		ImageUrl:      req.ImageUrl,
		CreatedBy:     userID,
	}
	if err := d.nftRepo.Upsert(ctx, nft); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to create nft set: %v", err)
		return nil, errorx.Unknown
	}

	nftMintHistory := &entity.NonFungibleTokenMintHistory{
		NonFungibleTokenID: nft.ID,
		Count:              int(req.Amount),
	}

	if err := d.nftMintHistoryRepo.Create(ctx, nftMintHistory); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to create nft mint history: %v", err)
		return nil, errorx.Unknown
	}

	xcontext.WithCommitDBTransaction(ctx)

	return &model.CreateNFTResponse{}, nil
}

func (d *nftDomain) GetNFT(ctx context.Context, req *model.GetNFTRequest) (*model.GetNFTResponse, error) {
	nft, err := d.nftRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found nft")
		}

		xcontext.Logger(ctx).Errorf("Unable to create nft set: %v", err)
		return nil, errorx.Unknown
	}

	aggResult, err := d.nftMintHistoryRepo.AggregateByNftID(ctx, req.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to aggregate nft by id: %v", err)
		return nil, errorx.Unknown
	}
	return &model.GetNFTResponse{
		Title:       nft.Title,
		Description: nft.Description,
		ImageUrl:    nft.ImageUrl,
		Chain:       nft.Chain,
		CreatedBy:   nft.CreatedBy,

		PendingAmount: aggResult.PendingAmount,
		ActiveAmount:  aggResult.ActiveAmount,
		FailureAmount: aggResult.FailureAmount,
	}, nil
}
