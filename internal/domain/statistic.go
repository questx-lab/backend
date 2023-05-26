package domain

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
	"github.com/redis/go-redis/v9"
)

type StatisticDomain interface {
	GetLeaderBoard(context.Context, *model.GetLeaderBoardRequest) (*model.GetLeaderBoardResponse, error)
}

type statisticDomain struct {
	claimedQuestRepo repository.ClaimedQuestRepository
	followerRepo     repository.FollowerRepository
	userRepo         repository.UserRepository
	redisClient      xredis.Client
}

func NewStatisticDomain(
	claimedQuestRepo repository.ClaimedQuestRepository,
	followerRepo repository.FollowerRepository,
	userRepo repository.UserRepository,
	redisClient xredis.Client,
) StatisticDomain {
	return &statisticDomain{
		claimedQuestRepo: claimedQuestRepo,
		followerRepo:     followerRepo,
		userRepo:         userRepo,
		redisClient:      redisClient,
	}
}

func (d *statisticDomain) GetLeaderBoard(
	ctx context.Context, req *model.GetLeaderBoardRequest,
) (*model.GetLeaderBoardResponse, error) {
	if req.CommunityID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow an empty community id")
	}

	if req.Limit == 0 {
		req.Limit = xcontext.Configs(ctx).ApiServer.DefaultLimit
	}

	if req.Limit < 0 {
		return nil, errorx.New(errorx.BadRequest, "Expected a positive limit")
	}

	if req.Limit > xcontext.Configs(ctx).ApiServer.MaxLimit {
		return nil, errorx.New(errorx.BadRequest, "Exceed the max limit")
	}

	period, err := stringToPeriod(req.Period)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid period: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid period")
	}

	leaderboard, err := d.getLeaderBoard(
		ctx, req.CommunityID, req.OrderedBy, period, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	lastPeriod, err := stringToLastPeriod(req.Period)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid period: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid period")
	}

	for i, info := range leaderboard {
		prevRank, err := d.getRank(
			ctx, info.User.ID, req.CommunityID, req.OrderedBy, lastPeriod)
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
			ID:           user.ID,
			Address:      user.Address.String,
			Name:         user.Name,
			Role:         string(user.Role),
			ReferralCode: user.ReferralCode,
			AvatarURL:    user.ProfilePicture,
			IsNewUser:    user.IsNewUser,
		}
	}

	return &model.GetLeaderBoardResponse{LeaderBoard: leaderboard}, nil
}

func (d *statisticDomain) getLeaderBoard(
	ctx context.Context,
	communityID string,
	orderedBy string,
	period entity.LeaderBoardPeriodType,
	offset, limit int,
) ([]model.UserStatistic, error) {
	key, err := redisKeyLeaderBoard(orderedBy, communityID, period)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid ordered by field: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid ordered by field")
	}

	ok, err := d.redisClient.Exist(ctx, key)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot call exist redis: %v", err)
		return nil, errorx.Unknown
	}

	// If the key didn't exist in redis, load it from database.
	if !ok {
		if err := d.loadLeaderboardFromDB(ctx, communityID, period); err != nil {
			return nil, err
		}
	}

	results, err := d.redisClient.ZRevRangeWithScores(ctx, key, offset, limit)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get revrange redis: %v", err)
		return nil, errorx.Unknown
	}

	leaderboard := []model.UserStatistic{}
	for i, z := range results {
		leaderboard = append(leaderboard, model.UserStatistic{
			User:        model.User{ID: z.Member.(string)},
			Value:       int(z.Score),
			CurrentRank: offset + i + 1,
		})
	}

	return leaderboard, nil
}

func (d *statisticDomain) getRank(
	ctx context.Context,
	userID string,
	communityID string,
	orderedBy string,
	period entity.LeaderBoardPeriodType,
) (uint64, error) {
	key, err := redisKeyLeaderBoard(orderedBy, communityID, period)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid ordered by field: %v", err)
		return 0, errorx.New(errorx.BadRequest, "Invalid ordered by field")
	}

	ok, err := d.redisClient.Exist(ctx, key)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot call exist redis: %v", err)
		return 0, errorx.Unknown
	}

	// If the key didn't exist in redis, load it from database.
	if !ok {
		if err := d.loadLeaderboardFromDB(ctx, communityID, period); err != nil {
			return 0, err
		}
	}

	rank, err := d.redisClient.ZRevRank(ctx, key, userID)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot get rev rank redis: %v", err)
		return 0, nil
	}

	return rank + 1, nil
}

func (d *statisticDomain) loadLeaderboardFromDB(
	ctx context.Context, communityID string, period entity.LeaderBoardPeriodType,
) error {
	followers, err := d.claimedQuestRepo.Statistic(
		ctx,
		repository.StatisticClaimedQuestFilter{
			CommunityID:   communityID,
			ReviewedStart: period.Start(),
			ReviewedEnd:   period.End(),
			Status:        []entity.ClaimedQuestStatus{entity.Accepted, entity.AutoAccepted},
		},
	)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot load statistic from database: %v", err)
		return errorx.Unknown
	}

	pointKey := redisKeyPointLeaderBoard(communityID, period)
	questKey := redisKeyQuestLeaderBoard(communityID, period)
	for _, f := range followers {
		fmt.Println(f.UserID, f.Points, f.Quests)
		err := d.redisClient.ZAdd(ctx, pointKey, redis.Z{Member: f.UserID, Score: float64(f.Points)})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot zadd redis: %v", err)
			return errorx.Unknown
		}

		err = d.redisClient.ZAdd(ctx, questKey, redis.Z{Member: f.UserID, Score: float64(f.Quests)})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot zadd redis: %v", err)
			return errorx.Unknown
		}
	}

	return nil
}
