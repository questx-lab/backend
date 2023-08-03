package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func Test_statisticDomain_GetLeaderboard(t *testing.T) {
	domain := NewStatisticDomain(
		repository.NewClaimedQuestRepository(),
		repository.NewFollowerRepository(),
		repository.NewUserRepository(&testutil.MockRedisClient{}),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
		statistic.New(
			repository.NewClaimedQuestRepository(),
			&testutil.MockRedisClient{
				ExistFunc: func(ctx context.Context, key string) (bool, error) {
					return true, nil
				},
				ZRevRangeWithScoresFunc: func(ctx context.Context, key string, offset, limit int) ([]redis.Z, error) {
					return []redis.Z{{Member: "user1", Score: 10}, {Member: "user2", Score: 8}}, nil
				},
				ZRevRankFunc: func(ctx context.Context, key, member string) (uint64, error) {
					if member == "user1" {
						return 1, nil
					}

					if member == "user2" {
						return 0, nil
					}

					return 10, nil
				},
			}),
	)

	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)
	resp, err := domain.GetLeaderBoard(ctx, &model.GetLeaderBoardRequest{
		Period:          "week",
		OrderedBy:       "point",
		CommunityHandle: testutil.Community1.Handle,
		Offset:          0,
		Limit:           2,
	})
	require.NoError(t, err)
	require.Equal(t, resp, &model.GetLeaderBoardResponse{
		LeaderBoard: []model.UserStatistic{
			{
				User: model.ShortUser{
					ID:        testutil.User1.ID,
					Name:      testutil.User1.Name,
					AvatarURL: testutil.User1.ProfilePicture,
				},
				Value:        10,
				CurrentRank:  1,
				PreviousRank: 2,
			},
			{
				User: model.ShortUser{
					ID:        testutil.User2.ID,
					Name:      testutil.User2.Name,
					AvatarURL: testutil.User2.ProfilePicture,
				},
				Value:        8,
				CurrentRank:  2,
				PreviousRank: 1,
			},
		},
	})
}
