package domain

import (
	"context"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type PayRewardDomain interface {
	GetMyPayRewards(context.Context, *model.GetMyPayRewardRequest) (*model.GetMyPayRewardResponse, error)
	GetClaimableRewards(context.Context, *model.GetClaimableRewardsRequest) (*model.GetClaimableRewardsResponse, error)
}

type payRewardDomain struct {
	payRewardRepo  repository.PayRewardRepository
	blockchainRepo repository.BlockChainRepository
	communityRepo  repository.CommunityRepository
	lotteryRepo    repository.LotteryRepository
	questFactory   questclaim.Factory
}

func NewPayRewardDomain(
	payRewardRepo repository.PayRewardRepository,
	blockchainRepo repository.BlockChainRepository,
	communityRepo repository.CommunityRepository,
	lotteryRepo repository.LotteryRepository,
	questFactory questclaim.Factory,
) *payRewardDomain {
	return &payRewardDomain{
		blockchainRepo: blockchainRepo,
		payRewardRepo:  payRewardRepo,
		communityRepo:  communityRepo,
		lotteryRepo:    lotteryRepo,
		questFactory:   questFactory,
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
			blockchainTx, err = d.blockchainRepo.GetTransactionByID(ctx, tx.TransactionID.String)
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

		payRewards = append(payRewards, model.ConvertPayReward(
			&tx,
			model.ConvertBlockchainToken(token),
			model.ConvertShortUser(nil, ""),
			referralCommunityHandle,
			fromCommunityHandle,
			model.ConvertBlockchainTransaction(blockchainTx),
		))
	}

	return &model.GetMyPayRewardResponse{PayRewards: payRewards}, nil
}

func (d *payRewardDomain) GetClaimableRewards(
	ctx context.Context, req *model.GetClaimableRewardsRequest,
) (*model.GetClaimableRewardsResponse, error) {
	requestUserID := xcontext.RequestUserID(ctx)

	tokenMap := map[string]float64{}
	response := model.GetClaimableRewardsResponse{
		ReferralCommunities:  []model.Community{},
		LotteryWinners:       []model.LotteryWinner{},
		TotalClaimableTokens: []model.ClaimableTokenInfo{},
	}

	referralCommunityToken, err := d.blockchainRepo.GetToken(
		ctx,
		xcontext.Configs(ctx).Quest.InviteCommunityRewardChain,
		xcontext.Configs(ctx).Quest.InviteCommunityRewardTokenAddress,
	)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot not get referral community token: %v", err)
	} else {
		communities, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
			ReferredBy:     requestUserID,
			ReferralStatus: []entity.ReferralStatusType{entity.ReferralClaimable},
		})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get claimable referral communities: %v", err)
			return nil, errorx.Unknown
		}

		for _, c := range communities {
			response.ReferralCommunities = append(response.ReferralCommunities, model.ConvertCommunity(&c, 0))
			tokenMap[referralCommunityToken.ID] += xcontext.Configs(ctx).Quest.InviteCommunityRewardAmount
		}
	}

	lotteryWinners, err := d.lotteryRepo.GetNotClaimedWinnerByUserID(ctx, requestUserID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get not claimed winners: %v", err)
		return nil, errorx.Unknown
	}

	prizeMap := map[string]entity.LotteryPrize{}
	for _, w := range lotteryWinners {
		prizeMap[w.LotteryPrizeID] = entity.LotteryPrize{}
	}

	prizes, err := d.lotteryRepo.GetPrizesByIDs(ctx, common.MapKeys(prizeMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get map prizes: %v", err)
		return nil, errorx.Unknown
	}

	for _, p := range prizes {
		prizeMap[p.ID] = p
	}

	for _, w := range lotteryWinners {
		p, ok := prizeMap[w.LotteryPrizeID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found prize %s", w.LotteryPrizeID)
			return nil, errorx.Unknown
		}

		response.LotteryWinners = append(response.LotteryWinners,
			model.ConvertLotteryWinner(&w, model.ConvertLotteryPrize(&p), model.ConvertShortUser(nil, "")))

		for _, r := range p.Rewards {
			if r.Type == entity.CoinReward {
				tokenMap[r.Data["token_id"].(string)] += r.Data["amount"].(float64)
			}
		}
	}

	tokens, err := d.blockchainRepo.GetTokensByIDs(ctx, common.MapKeys(tokenMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get tokens: %v", err)
		return nil, errorx.Unknown
	}

	for _, t := range tokens {
		amount, ok := tokenMap[t.ID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found token %s", t.ID)
		}

		response.TotalClaimableTokens = append(response.TotalClaimableTokens,
			model.ClaimableTokenInfo{
				TokenID:      t.ID,
				TokenSymbol:  t.Symbol,
				TokenAddress: t.Address,
				Chain:        t.Chain,
				Amount:       amount,
			})
	}

	return &response, nil
}
