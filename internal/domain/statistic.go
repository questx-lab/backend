package domain

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type StatisticDomain interface {
	GetLeaderBoard(context.Context, *model.GetLeaderBoardRequest) (*model.GetLeaderBoardResponse, error)
}

type statisticDomain struct {
	claimedQuestRepo repository.ClaimedQuestRepository
	followerRepo     repository.FollowerRepository
	userRepo         repository.UserRepository
	communityRepo    repository.CommunityRepository
	leaderboard      statistic.Leaderboard
}

func NewStatisticDomain(
	claimedQuestRepo repository.ClaimedQuestRepository,
	followerRepo repository.FollowerRepository,
	userRepo repository.UserRepository,
	communityRepo repository.CommunityRepository,
	leaderboard statistic.Leaderboard,
) StatisticDomain {
	return &statisticDomain{
		claimedQuestRepo: claimedQuestRepo,
		followerRepo:     followerRepo,
		userRepo:         userRepo,
		communityRepo:    communityRepo,
		leaderboard:      leaderboard,
	}
}

func (d *statisticDomain) GetLeaderBoard(
	ctx context.Context, req *model.GetLeaderBoardRequest,
) (*model.GetLeaderBoardResponse, error) {
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow an empty community id")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	apiCfg := xcontext.Configs(ctx).ApiServer
	if req.Limit == 0 {
		req.Limit = apiCfg.DefaultLimit
	}

	if req.Limit == -1 {
		return nil, errorx.New(errorx.BadRequest, "Limit must be positive")
	}

	if req.Limit > apiCfg.MaxLimit {
		return nil, errorx.New(errorx.BadRequest, "Exceed the maximum of limit (%d)", apiCfg.MaxLimit)
	}

	period, err := statistic.ToPeriod(req.Period)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid period: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid period")
	}

	leaderboard, err := d.leaderboard.GetLeaderBoard(
		ctx, community.ID, req.OrderedBy, period, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	lastPeriod, err := statistic.ToLastPeriod(req.Period)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid period: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid period")
	}

	for i, info := range leaderboard {
		prevRank, err := d.leaderboard.GetRank(
			ctx, info.User.ID, community.ID, req.OrderedBy, lastPeriod)
		if err != nil {
			return nil, err
		}
		leaderboard[i].PreviousRank = int(prevRank)

		user, err := d.userRepo.GetByID(ctx, info.User.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get user info: %v", err)
			return nil, errorx.Unknown
		}

		leaderboard[i].User = model.User{
			ID:            user.ID,
			WalletAddress: user.WalletAddress.String,
			Name:          user.Name,
			Role:          string(user.Role),
			ReferralCode:  user.ReferralCode,
			AvatarURL:     user.ProfilePicture,
			IsNewUser:     user.IsNewUser,
		}
	}

	return &model.GetLeaderBoardResponse{LeaderBoard: leaderboard}, nil
}
