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
		ProjectID: testutil.Project1.ID,
		Type:      "task",
		Offset:    0,
		Limit:     5,
	})

	taskActual := taskResp.Data

	require.NoError(t, err)
	taskExpected := []model.Achievement{
		{
			UserID:    testutil.Achievement1.UserID,
			TotalTask: int64(testutil.Achievement1.TotalTask),
			TotalExp:  testutil.Achievement1.TotalExp,
		},
		{
			UserID:    testutil.Achievement2.UserID,
			TotalTask: int64(testutil.Achievement1.TotalTask),
			TotalExp:  testutil.Achievement2.TotalExp,
		},
		{
			UserID:    testutil.Achievement3.UserID,
			TotalTask: int64(testutil.Achievement1.TotalTask),
			TotalExp:  testutil.Achievement3.TotalExp,
		},
	}
	for i := 0; i < len(taskActual); i++ {
		require.True(t, reflectutil.PartialEqual(taskActual[i], taskExpected[i]))
	}

	expResp, err := domain.GetLeaderBoard(ctx, &model.GetLeaderBoardRequest{
		Range:     string(entity.AchievementRangeWeek),
		ProjectID: testutil.Project1.ID,
		Type:      "task",
		Offset:    0,
		Limit:     5,
	})

	expActual := expResp.Data

	require.NoError(t, err)
	expExpected := []model.Achievement{
		{
			UserID:    testutil.Achievement1.UserID,
			TotalTask: int64(testutil.Achievement1.TotalTask),
			TotalExp:  testutil.Achievement1.TotalExp,
		},
		{
			UserID:    testutil.Achievement2.UserID,
			TotalTask: int64(testutil.Achievement1.TotalTask),
			TotalExp:  testutil.Achievement2.TotalExp,
		},
		{
			UserID:    testutil.Achievement3.UserID,
			TotalTask: int64(testutil.Achievement1.TotalTask),
			TotalExp:  testutil.Achievement3.TotalExp,
		},
	}
	for i := 0; i < len(expActual); i++ {
		require.True(t, reflectutil.PartialEqual(expActual[i], expExpected[i]))
	}
}
