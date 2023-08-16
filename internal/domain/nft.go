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
	nftSetRepo            repository.NftSetRepository
	nftRepo               repository.NftRepository
	communityRepo         repository.CommunityRepository
}

func NewNftDomain(
	communityRoleVerifier *common.CommunityRoleVerifier,
	blockchainCaller client.BlockchainCaller,
	nftSetRepo repository.NftSetRepository,
	nftRepo repository.NftRepository,
	communityRepo repository.CommunityRepository,
) *nftDomain {
	return &nftDomain{
		communityRoleVerifier: communityRoleVerifier,
		blockchainCaller:      blockchainCaller,
		nftSetRepo:            nftSetRepo,
		nftRepo:               nftRepo,
		communityRepo:         communityRepo,
	}
}

func (d *nftDomain) CreateNFTs(ctx context.Context, req *model.CreateNftsRequest) (*model.CreateNftsResponse, error) {
	userID := xcontext.RequestUserID(ctx)

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.communityRoleVerifier.Verify(ctx, community.ID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)
	set := &entity.NFTSet{
		SnowFlakeBase: entity.SnowFlakeBase{ID: xcontext.SnowFlake(ctx).Generate().Int64()},
		CommunityID:   community.ID,
		Title:         req.Title,
		ImageUrl:      req.ImageUrl,
		Chain:         req.Chain,
		CreatedBy:     userID,
	}
	if err := d.nftSetRepo.Create(ctx, set); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to create nft set: %v", err)
		return nil, errorx.Unknown
	}

	nfts := make([]*entity.NFT, 0, req.Amount)

	for i := 0; i < int(req.Amount); i++ {
		nfts = append(nfts, &entity.NFT{
			SetID: set.ID,
		})
	}

	if err := d.nftRepo.BulkInsert(ctx, nfts); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to create nfts: %v", err)
		return nil, errorx.Unknown
	}

	xcontext.WithCommitDBTransaction(ctx)

	return &model.CreateNftsResponse{}, nil
}

func (d *nftDomain) GetNFTs(ctx context.Context, req *model.GetNftsRequest) (*model.GetNftsResponse, error) {
	panic("not implemented") // TODO: Implement
}
