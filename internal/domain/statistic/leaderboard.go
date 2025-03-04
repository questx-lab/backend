package statistic

import (
	"context"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
)

type Leaderboard interface {
	GetLeaderBoard(
		ctx context.Context,
		communityID, orderedBy string,
		period entity.LeaderBoardPeriodType,
		offset, limit int,
	) ([]model.UserStatistic, error)

	GetRank(
		ctx context.Context,
		userID, communityID, orderedBy string,
		period entity.LeaderBoardPeriodType,
	) (uint64, error)

	ChangeQuestLeaderboard(
		ctx context.Context,
		value int64,
		reviewedAt time.Time,
		userID, communityID string,
	) error

	ChangePointLeaderboard(
		ctx context.Context,
		value int64,
		reviewedAt time.Time,
		userID, communityID string,
	) error
}

type leaderboard struct {
	claimedQuestRepo repository.ClaimedQuestRepository
	redisClient      xredis.Client
}

func New(
	claimedQuestRepo repository.ClaimedQuestRepository,
	redisClient xredis.Client,
) *leaderboard {
	return &leaderboard{
		claimedQuestRepo: claimedQuestRepo,
		redisClient:      redisClient,
	}
}

func (l *leaderboard) GetLeaderBoard(
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

	ok, err := l.redisClient.Exist(ctx, key)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot call exist redis: %v", err)
		return nil, errorx.Unknown
	}

	// If the key didn't exist in redis, load it from database.
	if !ok {
		if err := l.loadLeaderboardFromDB(ctx, communityID, period); err != nil {
			return nil, err
		}
	}

	results, err := l.redisClient.ZRevRangeWithScores(ctx, key, offset, limit)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get revrange redis: %v", err)
		return nil, errorx.Unknown
	}

	leaderboard := []model.UserStatistic{}
	for i, z := range results {
		leaderboard = append(leaderboard, model.UserStatistic{
			User:        model.ShortUser{ID: z.Member.(string)},
			Value:       int(z.Score),
			CurrentRank: offset + i + 1,
		})
	}

	return leaderboard, nil
}

func (l *leaderboard) GetRank(
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

	ok, err := l.redisClient.Exist(ctx, key)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot call exist redis: %v", err)
		return 0, errorx.Unknown
	}

	// If the key didn't exist in redis, load it from database.
	if !ok {
		if err := l.loadLeaderboardFromDB(ctx, communityID, period); err != nil {
			return 0, err
		}
	}

	rank, err := l.redisClient.ZRevRank(ctx, key, userID)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot get rev rank redis: %v", err)
		return 0, nil
	}

	return rank + 1, nil
}

func (l *leaderboard) changeLeaderboard(
	ctx context.Context,
	value int64,
	userID, communityID string,
	orderedBy string,
	period entity.LeaderBoardPeriodType,
) error {
	key, err := redisKeyLeaderBoard(orderedBy, communityID, period)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid ordered by field: %v", err)
		return errorx.New(errorx.BadRequest, "Invalid ordered by field")
	}

	ok, err := l.redisClient.Exist(ctx, key)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot call exist redis: %v", err)
		return errorx.Unknown
	}

	// If the key didn't exist in redis, no need to update.
	if !ok {
		return nil
	}

	if err := l.redisClient.ZIncrBy(ctx, key, value, userID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot call ZIncrBy redis: %v", err)
	}

	return nil
}

func (l *leaderboard) ChangeQuestLeaderboard(
	ctx context.Context,
	value int64,
	reviewedAt time.Time,
	userID, communityID string,
) error {
	for _, p := range periodConst {
		period, err := ToPeriodWithTime(p, reviewedAt)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Invalid period: %v", err)
			return errorx.Unknown
		}

		err = l.changeLeaderboard(ctx, value, userID, communityID, "quest", period)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *leaderboard) ChangePointLeaderboard(
	ctx context.Context,
	value int64,
	reviewedAt time.Time,
	userID, communityID string,
) error {
	for _, p := range periodConst {
		period, err := ToPeriodWithTime(p, reviewedAt)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Invalid period: %v", err)
			return errorx.Unknown
		}

		err = l.changeLeaderboard(ctx, value, userID, communityID, "point", period)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *leaderboard) loadLeaderboardFromDB(
	ctx context.Context, communityID string, period entity.LeaderBoardPeriodType,
) error {
	claimedQuestStatistic, err := l.claimedQuestRepo.Statistic(
		ctx,
		repository.StatisticClaimedQuestFilter{
			CommunityID:   communityID,
			ReviewedStart: period.Start(),
			ReviewedEnd:   period.End(),
			Status:        []entity.ClaimedQuestStatus{entity.Accepted, entity.AutoAccepted},
		},
	)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot load statistic from claimed quest: %v", err)
		return errorx.Unknown
	}

	pointKey := redisKeyPointLeaderBoard(communityID, period)
	questKey := redisKeyQuestLeaderBoard(communityID, period)
	for _, f := range claimedQuestStatistic {
		if f.UserID == "" {
			continue
		}

		err := l.redisClient.ZIncrBy(ctx, pointKey, int64(f.Points), f.UserID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot zadd redis point key: %v", err)
			return errorx.Unknown
		}

		err = l.redisClient.ZIncrBy(ctx, questKey, int64(f.Quests), f.UserID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot zadd redis quest key: %v", err)
			return errorx.Unknown
		}
	}

	return nil
}
