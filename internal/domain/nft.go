package domain

import (
	"context"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/model"
)

type NftDomain interface {
	CreateNFTs(context.Context, *model.CreateNftsRequest) (*model.CreateNftsResponse, error)
	GetNFTs(context.Context, *model.GetNftsRequest) (*model.GetNftsResponse, error)
}

type nftDomain struct {
	communityRoleVerifier *common.CommunityRoleVerifier
	blockchainCaller      client.BlockchainCaller
}

func NewNftDomain(communityRoleVerifier *common.CommunityRoleVerifier,
	blockchainCaller client.BlockchainCaller,
) *nftDomain {
	return &nftDomain{
		communityRoleVerifier: communityRoleVerifier,
		blockchainCaller:      blockchainCaller,
	}
}

func (d *nftDomain) CreateNFTs(_ context.Context, _ *model.CreateNftsRequest) (*model.CreateNftsResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (d *nftDomain) GetNFTs(_ context.Context, _ *model.GetNftsRequest) (*model.GetNftsResponse, error) {
	panic("not implemented") // TODO: Implement
}
