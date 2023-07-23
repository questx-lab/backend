package domain

import (
	"context"
	"errors"
	"net/url"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/ethutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type BlockchainDomain interface {
	CreateChain(context.Context, *model.CreateBlockchainRequest) (*model.CreateBlockchainResponse, error)
	CreateConnection(context.Context, *model.CreateBlockchainConnectionRequest) (*model.CreateBlockchainConnectionResponse, error)
	DeleteConnection(context.Context, *model.DeleteBlockchainConnectionRequest) (*model.DeleteBlockchainConnectionResponse, error)
	GetWalletAddress(context.Context, *model.GetCommunityWalletAddressRequest) (*model.GetCommunityWalletAddressResponse, error)
	CreateToken(context.Context, *model.CreateBlockchainTokenRequest) (*model.CreateBlockchainTokenResponse, error)
}

type blockchainDomain struct {
	blockchainRepo repository.BlockChainRepository
	communityRepo  repository.CommunityRepository

	blockchainCaller client.BlockchainCaller
}

func NewBlockchainDomain(
	blockchainRepo repository.BlockChainRepository,
	communityRepo repository.CommunityRepository,
	blockchainCaller client.BlockchainCaller,
) *blockchainDomain {
	return &blockchainDomain{
		blockchainRepo:   blockchainRepo,
		communityRepo:    communityRepo,
		blockchainCaller: blockchainCaller,
	}
}

func (d *blockchainDomain) CreateChain(
	ctx context.Context, req *model.CreateBlockchainRequest,
) (*model.CreateBlockchainResponse, error) {
	err := d.blockchainRepo.Upsert(ctx, &entity.Blockchain{
		Name:                 req.Chain,
		ID:                   req.ChainID,
		UseExternalRPC:       req.UseExternalRPC,
		UseEip1559:           req.UseEip1559,
		BlockTime:            req.BlockTime,
		AdjustTime:           req.AdjustTime,
		ThresholdUpdateBlock: req.ThresholdUpdateBlock,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create block chain: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateBlockchainResponse{}, nil
}

func (d *blockchainDomain) CreateConnection(
	ctx context.Context, req *model.CreateBlockchainConnectionRequest,
) (*model.CreateBlockchainConnectionResponse, error) {
	typeEnum, err := enum.ToEnum[entity.BlockchainConnectionType](req.Type)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid type: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid type")
	}

	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid URL: %v", err)
		return nil, errorx.Unknown
	}

	if parsedURL.Scheme != "" {
		return nil, errorx.New(errorx.BadRequest, "Do not include scheme into url")
	}

	err = d.blockchainRepo.CreateBlockchainConnection(ctx, &entity.BlockchainConnection{
		Chain: req.Chain,
		Type:  typeEnum,
		URL:   req.URL,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create connection: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateBlockchainConnectionResponse{}, nil
}

func (d *blockchainDomain) DeleteConnection(
	ctx context.Context, req *model.DeleteBlockchainConnectionRequest,
) (*model.DeleteBlockchainConnectionResponse, error) {
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid URL: %v", err)
		return nil, errorx.Unknown
	}

	if parsedURL.Scheme != "" {
		return nil, errorx.New(errorx.BadRequest, "Do not include scheme into url")
	}

	err = d.blockchainRepo.DeleteBlockchainConnection(ctx, req.Chain, req.URL)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete connection: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteBlockchainConnectionResponse{}, nil
}

func (d *blockchainDomain) GetWalletAddress(
	ctx context.Context, req *model.GetCommunityWalletAddressRequest,
) (*model.GetCommunityWalletAddressResponse, error) {
	// Everyone can get wallet address of any community or our platform.
	walletNonce := "" // Our platform use an empty nonce.
	if req.CommunityHandle != "" {
		community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found community")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}

		if community.WalletNonce == "" {
			nonce, err := crypto.GenerateRandomString()
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot generate nonce: %v", err)
				return nil, errorx.Unknown
			}

			err = d.communityRepo.UpdateByID(ctx, community.ID, entity.Community{
				WalletNonce: nonce,
			})
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot update wallet nonce: %v", err)
				return nil, errorx.Unknown
			}

			community.WalletNonce = nonce
		}

		walletNonce = community.WalletNonce
	}

	walletPrivateKey, err := ethutil.GeneratePrivateKey(
		[]byte(xcontext.Configs(ctx).Blockchain.SecretKey), []byte(walletNonce))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get wallet address: %v", err)
		return nil, errorx.Unknown
	}

	publicKey := ethcrypto.PubkeyToAddress(walletPrivateKey.PublicKey)
	return &model.GetCommunityWalletAddressResponse{WalletAddress: publicKey.String()}, nil
}

func (d *blockchainDomain) CreateToken(
	ctx context.Context, req *model.CreateBlockchainTokenRequest,
) (*model.CreateBlockchainTokenResponse, error) {
	if req.Chain == "" {
		return nil, errorx.New(errorx.BadRequest, "Require chain")
	}

	if req.Address == "" {
		return nil, errorx.New(errorx.BadRequest, "Require address")
	}

	info, err := d.blockchainCaller.GetTokenInfo(ctx, req.Chain, req.Address)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get token info: %v", err)
		return nil, errorx.Unknown
	}

	err = d.blockchainRepo.CreateToken(ctx, &entity.BlockchainToken{
		Base:     entity.Base{ID: uuid.NewString()},
		Chain:    req.Chain,
		Address:  req.Address,
		Symbol:   info.Symbol,
		Name:     info.Name,
		Decimals: info.Decimals,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create token: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateBlockchainTokenResponse{
		Token:    info.Symbol,
		Name:     info.Name,
		Decimals: info.Decimals,
	}, nil
}
