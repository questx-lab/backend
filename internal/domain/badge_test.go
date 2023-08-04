package domain

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
)

func Test_badgeDomain_FollowCommunity_and_GetMyBadges(t *testing.T) {
	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)

	userRepo := repository.NewUserRepository(&testutil.MockRedisClient{})
	oauth2Repo := repository.NewOAuth2Repository()
	followerRepo := repository.NewFollowerRepository()
	followerRoleRepo := repository.NewFollowerRoleRepository()
	badgeRepo := repository.NewBadgeRepository()
	badgeDetailRepo := repository.NewBadgeDetailRepository()
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{})
	claimedQuestRepo := repository.NewClaimedQuestRepository()

	newUser := &entity.User{Base: entity.Base{ID: uuid.NewString()}}
	require.NoError(t, userRepo.Create(ctx, newUser))

	userDomain := NewUserDomain(
		userRepo,
		oauth2Repo,
		followerRepo,
		followerRoleRepo,
		communityRepo,
		claimedQuestRepo,
		badge.NewManager(
			badgeRepo,
			badgeDetailRepo,
			&testutil.MockBadge{
				NameValue:     badge.SharpScoutBadgeName,
				IsGlobalValue: false,
				ScanFunc: func(ctx context.Context, userID, communityID string) ([]entity.Badge, error) {
					return []entity.Badge{testutil.BadgeSharpScout1}, nil
				},
			},
		),
		nil, nil, &testutil.MockRedisClient{},
	)

	ctx = xcontext.WithRequestUserID(ctx, newUser.ID)
	_, err := userDomain.FollowCommunity(ctx, &model.FollowCommunityRequest{
		CommunityHandle: testutil.Community1.Handle,
		InvitedBy:       testutil.Follower1.UserID,
	})
	require.NoError(t, err)

	badgeDomain := NewBadgeDomain(
		badgeRepo, badgeDetailRepo, communityRepo,
		badge.NewManager(
			badgeRepo, badgeDetailRepo,
			&testutil.MockBadge{NameValue: badge.QuestWarriorBadgeName},
		),
	)

	// Get badges and check their level, name. Ensure that they haven't been
	// notified to client yet.
	ctx = xcontext.WithRequestUserID(ctx, testutil.Follower1.UserID)
	badges, err := badgeDomain.GetMyBadgeDetails(ctx, &model.GetMyBadgeDetailsRequest{
		CommunityHandle: testutil.Community1.Handle,
	})
	require.NoError(t, err)
	require.Len(t, badges.BadgeDetails, 1)
	require.Equal(t, testutil.BadgeSharpScout1.ID, badges.BadgeDetails[0].Badge.ID)
	require.False(t, badges.BadgeDetails[0].WasNotified)

	// Get badges again and ensure they was notified to client.
	badges, err = badgeDomain.GetMyBadgeDetails(ctx, &model.GetMyBadgeDetailsRequest{
		CommunityHandle: testutil.Community1.Handle,
	})
	require.NoError(t, err)
	require.Len(t, badges.BadgeDetails, 1)
	require.Equal(t, testutil.BadgeSharpScout1.ID, badges.BadgeDetails[0].Badge.ID)
	require.True(t, badges.BadgeDetails[0].WasNotified)
}
