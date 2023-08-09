package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
)

func Test_statisticDomain_GetLeaderboard(t *testing.T) {
	ctx := testutil.MockContext(t)
	testutil.CreateFixtureDb(ctx)

	domain := NewStatisticDomain(
		repository.NewClaimedQuestRepository(),
		repository.NewFollowerRepository(),
		repository.NewUserRepository(testutil.RedisClient(ctx)),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(ctx)),
		statistic.New(
			repository.NewClaimedQuestRepository(),
			testutil.RedisClient(ctx),
		),
	)

	claimedQuestDomain := NewClaimedQuestDomain(
		repository.NewClaimedQuestRepository(),
		repository.NewQuestRepository(&testutil.MockSearchCaller{}),
		repository.NewFollowerRepository(),
		repository.NewFollowerRoleRepository(),
		repository.NewUserRepository(testutil.RedisClient(ctx)),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(ctx)),
		repository.NewCategoryRepository(),
		badge.NewManager(repository.NewBadgeRepository(),
			repository.NewBadgeDetailRepository(),
			&testutil.MockBadge{NameValue: badge.SharpScoutBadgeName},
			&testutil.MockBadge{NameValue: badge.RainBowBadgeName},
			&testutil.MockBadge{NameValue: badge.QuestWarriorBadgeName},
		),
		statistic.New(
			repository.NewClaimedQuestRepository(),
			testutil.RedisClient(ctx),
		),
		testutil.CommunityRoleVerifier,
		nil, testutil.QuestFactory, testutil.RedisClient(ctx),
	)

	_, err := claimedQuestDomain.Claim(
		xcontext.WithRequestUserID(ctx, testutil.User1.ID),
		&model.ClaimQuestRequest{
			QuestID: testutil.Quest3.ID,
		},
	)
	require.NoError(t, err)

	_, err = claimedQuestDomain.Claim(
		xcontext.WithRequestUserID(ctx, testutil.User2.ID),
		&model.ClaimQuestRequest{
			QuestID: testutil.Quest4.ID,
		},
	)
	require.NoError(t, err)

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
				Value:       100,
				CurrentRank: 1,
			},
			{
				User: model.ShortUser{
					ID:        testutil.User2.ID,
					Name:      testutil.User2.Name,
					AvatarURL: testutil.User2.ProfilePicture,
				},
				Value:       80,
				CurrentRank: 2,
			},
		},
	})
}
