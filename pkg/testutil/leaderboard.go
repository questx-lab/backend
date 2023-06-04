package testutil

import (
	"context"

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

	IncreaseLeaderboardFunc func(
		ctx context.Context,
		value uint64,
		userID, communityID, orderedBy string,
		period entity.LeaderBoardPeriodType,
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

func (m *MockLeaderboard) IncreaseLeaderboard(
	ctx context.Context,
	value uint64,
	userID, communityID, orderedBy string,
	period entity.LeaderBoardPeriodType,
) error {
	if m.IncreaseLeaderboardFunc != nil {
		return m.IncreaseLeaderboardFunc(ctx, value, userID, communityID, orderedBy, period)
	}

	return nil
}
