package domain

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type PayRewardDomain interface {
	GetMyPayRewards(context.Context, *model.GetMyPayRewardRequest) (*model.GetMyPayRewardResponse, error)
}

type payRewardDomain struct {
	payRewardRepo  repository.PayRewardRepository
	blockchainRepo repository.BlockChainRepository
	communityRepo  repository.CommunityRepository
}

func NewPayRewardDomain(
	payRewardRepo repository.PayRewardRepository,
	blockchainRepo repository.BlockChainRepository,
	communityRepo repository.CommunityRepository,
) *payRewardDomain {
	return &payRewardDomain{
		blockchainRepo: blockchainRepo,
		payRewardRepo:  payRewardRepo,
		communityRepo:  communityRepo,
	}
}

func (d *payRewardDomain) GetMyPayRewards(
	ctx context.Context, req *model.GetMyPayRewardRequest,
) (*model.GetMyPayRewardResponse, error) {
	txs, err := d.payRewardRepo.GetByUserID(ctx, xcontext.RequestUserID(ctx))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get payReward by user id: %v", err)
		return nil, errorx.Unknown
	}

	payRewards := []model.PayReward{}
	for _, tx := range txs {
		var blockchainTx *entity.BlockchainTransaction
		if tx.TransactionID.Valid {
			var err error
			blockchainTx, err = d.blockchainRepo.GetByID(ctx, tx.TransactionID.String)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get blockchain transaction by id: %v", err)
				return nil, errorx.Unknown
			}
		}

		var referralCommunityHandle string
		if tx.ReferralCommunityID.Valid {
			community, err := d.communityRepo.GetByID(ctx, tx.ReferralCommunityID.String)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get referral community: %v", err)
				return nil, errorx.Unknown
			}

			referralCommunityHandle = community.Handle
		}

		var fromCommunityHandle string
		if tx.FromCommunityID.Valid {
			community, err := d.communityRepo.GetByID(ctx, tx.FromCommunityID.String)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get from community: %v", err)
				return nil, errorx.Unknown
			}

			fromCommunityHandle = community.Handle
		}

		token, err := d.blockchainRepo.GetTokenByID(ctx, tx.TokenID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get token by id: %v", err)
			return nil, errorx.Unknown
		}

		payRewards = append(payRewards, convertPayReward(
			&tx,
			convertBlockchainToken(token),
			convertUser(nil, nil, false),
			referralCommunityHandle,
			fromCommunityHandle,
			convertBlockchainTransaction(blockchainTx),
		))
	}

	return &model.GetMyPayRewardResponse{PayRewards: payRewards}, nil
}
