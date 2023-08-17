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

type NftDomain interface {
	CreateNFTs(context.Context, *model.CreateNftsRequest) (*model.CreateNftsResponse, error)
	GetNFTs(context.Context, *model.GetNftsRequest) (*model.GetNftsResponse, error)
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

func (d *nftDomain) CreateNFTs(ctx context.Context, req *model.CreateNftsRequest) (*model.CreateNftsResponse, error) {
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

	return &model.CreateNftsResponse{}, nil
}

func (d *nftDomain) GetNFTs(ctx context.Context, req *model.GetNftsRequest) (*model.GetNftsResponse, error) {
	panic("not implemented") // TODO: Implement
}
