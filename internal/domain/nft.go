package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/pinata"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type NFTDomain interface {
	CreateNFT(context.Context, *model.CreateNFTRequest) (*model.CreateNFTResponse, error)
	GetNFT(context.Context, *model.GetNFTRequest) (*model.GetNFTResponse, error)
	GetNFTs(context.Context, *model.GetNFTsRequest) (*model.GetNFTsResponse, error)
	GetNFTsByCommunity(context.Context, *model.GetNFTsByCommunityRequest) (*model.GetNFTsByCommunityResponse, error)
	GetMyNFTs(context.Context, *model.GetMyNFTsRequest) (*model.GetMyNFTsResponse, error)
}

type nftDomain struct {
	communityRoleVerifier *common.CommunityRoleVerifier
	blockchainCaller      client.BlockchainCaller
	nftRepo               repository.NftRepository
	communityRepo         repository.CommunityRepository

	pinataEndpoint pinata.IEndpoint
}

func NewNftDomain(
	communityRoleVerifier *common.CommunityRoleVerifier,
	blockchainCaller client.BlockchainCaller,
	nftRepo repository.NftRepository,
	communityRepo repository.CommunityRepository,
	pinataEndpoint pinata.IEndpoint,
) *nftDomain {
	return &nftDomain{
		communityRoleVerifier: communityRoleVerifier,
		blockchainCaller:      blockchainCaller,
		nftRepo:               nftRepo,
		communityRepo:         communityRepo,
		pinataEndpoint:        pinataEndpoint,
	}
}

func (d *nftDomain) CreateNFT(ctx context.Context, req *model.CreateNFTRequest) (*model.CreateNFTResponse, error) {
	if req.ID == 0 && req.Name == "" {
		return nil, errorx.New(errorx.BadRequest, "NFT needs a name")
	}

	if req.ID == 0 && req.Image == "" {
		return nil, errorx.New(errorx.BadRequest, "You need upload an image for this NFT")
	}

	if req.ID != 0 && req.Amount == 0 {
		return nil, errorx.New(errorx.BadRequest, "You need determine amount of NFT you want to mint")
	}

	var id int64
	var communityID string
	var ipfs string
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

		resp, err := xcontext.HTTPClient(ctx).Get(req.Image)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get the image: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Cannot determine the NFT image")
		}

		if resp.StatusCode != 200 {
			xcontext.Logger(ctx).Errorf("Invalid status code when get image: %d", resp.StatusCode)
			return nil, errorx.Unknown
		}

		imageHash, err := d.pinataEndpoint.PinFile(ctx, fmt.Sprintf("%d.image", id), resp.Body)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot push image to ipfs: %v", err)
			return nil, errorx.New(errorx.Unavailable, "Cannot push image to ipfs")
		}

		content := model.NonFungibleTokenContent{
			TokenID:    id,
			Name:       req.Name,
			Decription: req.Description,
			Image:      fmt.Sprintf("ipfs://%s", imageHash),
			Properties: model.NonFungibleTokenProperties{
				CommunityID: communityID,
			},
		}

		bContent, err := json.Marshal(content)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot marshal the nft content: %v", err)
			return nil, errorx.Unknown
		}

		nftHash, err := d.pinataEndpoint.PinFile(ctx, fmt.Sprintf("%d.json", id), bytes.NewBuffer(bContent))
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot push nft content to ipfs: %v", err)
			return nil, errorx.New(errorx.Unavailable, "Cannot push nft content to ipfs")
		}

		ipfs = fmt.Sprintf("ipfs://%s", nftHash)
		nft := &entity.NonFungibleToken{
			SnowFlakeBase: entity.SnowFlakeBase{ID: id},
			CommunityID:   community.ID,
			CreatedBy:     xcontext.RequestUserID(ctx),
			Chain:         req.Chain,
			Name:          req.Name,
			Description:   req.Description,
			Image:         req.Image,
			IpfsImage:     content.Image,
			Ipfs:          ipfs,
		}
		if err := d.nftRepo.Create(ctx, nft); err != nil {
			xcontext.Logger(ctx).Errorf("Unable to create nft set: %v", err)
			return nil, errorx.Unknown
		}
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
		ipfs = nft.IpfsImage
		communityID = nft.CommunityID
	}

	if err := d.communityRoleVerifier.Verify(ctx, communityID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	if req.Amount > 0 {
		txID, err := d.blockchainCaller.MintNFT(ctx, communityID, req.Chain, id, req.Amount, ipfs)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Unable to mint nft: %v", err)
			return nil, errorx.Unknown
		}

		nftMintHistory := &entity.NonFungibleTokenMintHistory{
			NonFungibleTokenID: id,
			TransactionID:      txID,
			Amount:             int(req.Amount),
		}

		if err := d.nftRepo.CreateHistory(ctx, nftMintHistory); err != nil {
			xcontext.Logger(ctx).Errorf("Unable to create nft mint history: %v", err)
			return nil, errorx.Unknown
		}
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.CreateNFTResponse{ID: id}, nil
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

	return &model.GetNFTResponse{NFT: model.ConvertNFT(nft)}, nil
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
		result = append(result, model.ConvertNFT(&nft))
	}

	return &model.GetNFTsByCommunityResponse{NFTs: result}, nil
}

func (d *nftDomain) GetNFTs(ctx context.Context, req *model.GetNFTsRequest) (*model.GetNFTsResponse, error) {
	nftIDs := []int64{}
	parts := strings.Split(req.NftIDs, ",")
	for _, p := range parts {
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Cannot convert string to int: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid id %s", p)
		}

		nftIDs = append(nftIDs, id)
	}

	nfts, err := d.nftRepo.GetByIDs(ctx, nftIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found nft")
		}

		xcontext.Logger(ctx).Errorf("Unable to create nft set: %v", err)
		return nil, errorx.Unknown
	}

	result := []model.NonFungibleToken{}
	for _, nft := range nfts {
		result = append(result, model.ConvertNFT(&nft))
	}

	return &model.GetNFTsResponse{NFTs: result}, nil
}

func (d *nftDomain) GetMyNFTs(
	ctx context.Context, req *model.GetMyNFTsRequest,
) (*model.GetMyNFTsResponse, error) {
	claimedNFTs, err := d.nftRepo.GetByUserID(ctx, xcontext.RequestUserID(ctx))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get claimed nft of user: %v", err)
		return nil, errorx.Unknown
	}

	nftIDs := []int64{}
	for _, claimedNFT := range claimedNFTs {
		nftIDs = append(nftIDs, claimedNFT.NonFungibleTokenID)
	}

	nfts, err := d.nftRepo.GetByIDs(ctx, nftIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get nfts info: %v", err)
		return nil, errorx.Unknown
	}

	nftMap := map[int64]entity.NonFungibleToken{}
	for i := range nfts {
		nftMap[nfts[i].ID] = nfts[i]
	}

	userNFTs := []model.UserNonFungibleToken{}
	for _, claimedNFT := range claimedNFTs {
		nft, ok := nftMap[claimedNFT.NonFungibleTokenID]
		if !ok {
			xcontext.Logger(ctx).Warnf(
				"Not found nft %s of %s", claimedNFT.NonFungibleTokenID, claimedNFT.UserID)
			continue
		}

		userNFTs = append(userNFTs, model.ConvertUserNFT(&claimedNFT, model.ConvertNFT(&nft)))
	}

	return &model.GetMyNFTsResponse{NFTs: userNFTs}, nil
}
