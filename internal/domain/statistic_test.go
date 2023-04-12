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

	achievementRepo := repository.NewAchievementRepository()
	domain := NewStatisticDomain(achievementRepo)

	taskResp, err := domain.GetLeaderBoard(ctx, &model.GetLeaderBoardRequest{
		Range:     string(entity.AchievementRangeWeek),
		ProjectID: testutil.Project2.ID,
		Type:      "task",
		Offset:    0,
		Limit:     5,
	})
	require.NoError(t, err)

	taskActual := taskResp.Data

	taskExpected := []model.Achievement{
		{
			UserID:     testutil.Achievement1.UserID,
			TotalTask:  testutil.Achievement1.TotalTask,
			TotalPoint: testutil.Achievement1.TotalPoint,
		},
		{
			UserID:     testutil.Achievement2.UserID,
			TotalTask:  testutil.Achievement2.TotalTask,
			TotalPoint: testutil.Achievement2.TotalPoint,
		},
		{
			UserID:     testutil.Achievement3.UserID,
			TotalTask:  testutil.Achievement3.TotalTask,
			TotalPoint: testutil.Achievement3.TotalPoint,
		},
	}

	require.Equal(t, len(taskExpected), len(taskActual))
	for i := 0; i < len(taskActual); i++ {
		require.True(t, reflectutil.PartialEqual(&taskExpected[i], &taskActual[i]))
	}

	expResp, err := domain.GetLeaderBoard(ctx, &model.GetLeaderBoardRequest{
		Range:     string(entity.AchievementRangeWeek),
		ProjectID: testutil.Project2.ID,
		Type:      "point",
		Offset:    0,
		Limit:     5,
	})
	require.NoError(t, err)

	expActual := expResp.Data

	expExpected := []model.Achievement{
		{
			UserID:     testutil.Achievement3.UserID,
			TotalTask:  testutil.Achievement3.TotalTask,
			TotalPoint: testutil.Achievement3.TotalPoint,
		},
		{
			UserID:     testutil.Achievement2.UserID,
			TotalTask:  testutil.Achievement2.TotalTask,
			TotalPoint: testutil.Achievement2.TotalPoint,
		},
		{
			UserID:     testutil.Achievement1.UserID,
			TotalTask:  testutil.Achievement1.TotalTask,
			TotalPoint: testutil.Achievement1.TotalPoint,
		},
	}
	require.Equal(t, len(expExpected), len(expActual))
	for i := 0; i < len(expExpected); i++ {
		require.True(t, reflectutil.PartialEqual(&expExpected[i], &expActual[i]))
	}
}
