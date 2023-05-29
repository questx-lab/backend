package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/domain/leaderboard"
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
		repository.NewUserRepository(),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
		leaderboard.New(repository.NewClaimedQuestRepository(), &testutil.MockRedisClient{
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
				User: model.User{
					ID:      "user1",
					Address: "",
					Name:    "user1",
					Role:    "super_admin",
				},
				Value:        10,
				CurrentRank:  1,
				PreviousRank: 2,
			},
			{
				User: model.User{
					ID:      "user2",
					Address: "random-wallet-address",
					Name:    "user2",
					Role:    "",
				},
				Value:        8,
				CurrentRank:  2,
				PreviousRank: 1,
			},
		},
	})
}
