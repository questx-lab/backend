package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/questx-lab/backend/pkg/testutil"

	"github.com/stretchr/testify/require"
)

func Test_statisticDomain_GetLeaderBoard(t *testing.T) {
	ctx := testutil.NewMockContextWithUserID(nil, testutil.Project1.CreatedBy)
	testutil.CreateFixtureDb(ctx)

	domain := NewStatisticDomain(
		repository.NewUserAggregateRepository(),
		repository.NewUserRepository(),
	)

	taskResp, err := domain.GetLeaderBoard(ctx, &model.GetLeaderBoardRequest{
		Range:     string(entity.UserAggregateRangeWeek),
		ProjectID: testutil.Project2.ID,
		Type:      "task",
		Offset:    0,
		Limit:     5,
	})
	require.NoError(t, err)

	taskActual := taskResp.LeaderBoard

	taskExpected := []model.UserAggregate{
		{
			UserID:      testutil.UserAggregate3.UserID,
			TotalTask:   testutil.UserAggregate3.TotalTask,
			TotalPoint:  testutil.UserAggregate3.TotalPoint,
			PrevRank:    3,
			CurrentRank: 1,
		},
		{
			UserID:      testutil.UserAggregate2.UserID,
			TotalTask:   testutil.UserAggregate2.TotalTask,
			TotalPoint:  testutil.UserAggregate2.TotalPoint,
			PrevRank:    1,
			CurrentRank: 2,
		},
		{
			UserID:      testutil.UserAggregate1.UserID,
			TotalTask:   testutil.UserAggregate1.TotalTask,
			TotalPoint:  testutil.UserAggregate1.TotalPoint,
			PrevRank:    2,
			CurrentRank: 3,
		},
	}

	require.Equal(t, len(taskExpected), len(taskActual))
	for i := 0; i < len(taskActual); i++ {
		require.True(t, reflectutil.PartialEqual(&taskExpected[i], &taskActual[i]))
	}

	expResp, err := domain.GetLeaderBoard(ctx, &model.GetLeaderBoardRequest{
		Range:     string(entity.UserAggregateRangeWeek),
		ProjectID: testutil.Project2.ID,
		Type:      "point",
		Offset:    0,
		Limit:     5,
	})
	require.NoError(t, err)

	expActual := expResp.LeaderBoard

	expExpected := []model.UserAggregate{
		{
			UserID:      testutil.UserAggregate1.UserID,
			TotalTask:   testutil.UserAggregate1.TotalTask,
			TotalPoint:  testutil.UserAggregate1.TotalPoint,
			PrevRank:    1,
			CurrentRank: 1,
		},
		{
			UserID:      testutil.UserAggregate2.UserID,
			TotalTask:   testutil.UserAggregate2.TotalTask,
			TotalPoint:  testutil.UserAggregate2.TotalPoint,
			PrevRank:    2,
			CurrentRank: 2,
		},
		{
			UserID:      testutil.UserAggregate3.UserID,
			TotalTask:   testutil.UserAggregate3.TotalTask,
			TotalPoint:  testutil.UserAggregate3.TotalPoint,
			PrevRank:    3,
			CurrentRank: 3,
		},
	}
	require.Equal(t, len(expExpected), len(expActual))
	for i := 0; i < len(expExpected); i++ {
		require.True(t, reflectutil.PartialEqual(&expExpected[i], &expActual[i]))
	}
}
