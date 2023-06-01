package testutil

import (
	"context"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
)

type MockLeaderboard struct {
	GetLeaderBoardFunc func(
		ctx context.Context,
		communityID, orderedBy string,
		period entity.LeaderBoardPeriodType,
		offset, limit int,
	) ([]model.UserStatistic, error)

	GetRankFunc func(
		ctx context.Context,
		userID, communityID, orderedBy string,
		period entity.LeaderBoardPeriodType,
	) (uint64, error)

	ChangeQuestLeaderboardFunc func(
		ctx context.Context,
		value int64,
		reviewedAt time.Time,
		userID, communityID string,
	) error

	ChangePointLeaderboardFunc func(
		ctx context.Context,
		value int64,
		reviewedAt time.Time,
		userID, communityID string,
	) error
}

func (m *MockLeaderboard) GetLeaderBoard(
	ctx context.Context,
	communityID, orderedBy string,
	period entity.LeaderBoardPeriodType,
	offset, limit int,
) ([]model.UserStatistic, error) {
	if m.GetLeaderBoardFunc != nil {
		return m.GetLeaderBoardFunc(ctx, communityID, orderedBy, period, offset, limit)
	}

	return nil, nil
}

func (m *MockLeaderboard) GetRank(
	ctx context.Context,
	userID, communityID, orderedBy string,
	period entity.LeaderBoardPeriodType,
) (uint64, error) {
	if m.GetRankFunc != nil {
		return m.GetRankFunc(ctx, userID, communityID, orderedBy, period)
	}

	return 0, nil
}

func (m *MockLeaderboard) ChangeQuestLeaderboard(
	ctx context.Context,
	value int64,
	reviewedAt time.Time,
	userID, communityID string,
) error {
	if m.ChangeQuestLeaderboardFunc != nil {
		return m.ChangeQuestLeaderboardFunc(ctx, value, reviewedAt, userID, communityID)
	}

	return nil
}

func (m *MockLeaderboard) ChangePointLeaderboard(
	ctx context.Context,
	value int64,
	reviewedAt time.Time,
	userID, communityID string,
) error {
	if m.ChangePointLeaderboardFunc != nil {
		return m.ChangePointLeaderboardFunc(ctx, value, reviewedAt, userID, communityID)
	}

	return nil
}
