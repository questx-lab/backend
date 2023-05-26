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

func Test_userDomain_GetReferralInfo(t *testing.T) {
	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)

	domain := NewUserDomain(
		repository.NewUserRepository(),
		repository.NewOAuth2Repository(),
		repository.NewFollowerRepository(),
		repository.NewBadgeRepository(),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
		badge.NewManager(
			repository.NewBadgeRepository(),
			&testutil.MockBadge{
				NameValue:     badge.SharpScoutBadgeName,
				IsGlobalValue: false,
				ScanFunc: func(ctx context.Context, userID, communityID string) (int, error) {
					return 0, nil
				},
			},
		),
		nil,
	)

	inviteResp, err := domain.GetInvite(ctx, &model.GetInviteRequest{
		InviteCode: testutil.Follower1.InviteCode,
	})
	require.NoError(t, err)
	require.Equal(t, inviteResp.Community.ID, testutil.Community1.ID)
	require.Equal(t, inviteResp.Community.Handle, testutil.Community1.Handle)
}

func Test_userDomain_FollowCommunity_and_GetMyBadges(t *testing.T) {
	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)

	userRepo := repository.NewUserRepository()
	oauth2Repo := repository.NewOAuth2Repository()
	pariticipantRepo := repository.NewFollowerRepository()
	badgeRepo := repository.NewBadgeRepository()
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{})

	newUser := &entity.User{Base: entity.Base{ID: uuid.NewString()}}
	require.NoError(t, userRepo.Create(ctx, newUser))

	domain := NewUserDomain(
		userRepo,
		oauth2Repo,
		pariticipantRepo,
		badgeRepo,
		communityRepo,
		badge.NewManager(
			badgeRepo,
			&testutil.MockBadge{
				NameValue:     badge.SharpScoutBadgeName,
				IsGlobalValue: false,
				ScanFunc: func(ctx context.Context, userID, communityID string) (int, error) {
					return 1, nil
				},
			},
		),
		nil,
	)

	ctx = xcontext.WithRequestUserID(ctx, newUser.ID)
	_, err := domain.FollowCommunity(ctx, &model.FollowCommunityRequest{
		CommunityID: testutil.Follower1.CommunityID,
		InvitedBy:   testutil.Follower1.UserID,
	})
	require.NoError(t, err)

	// Get badges and check their level, name. Ensure that they haven't been
	// notified to client yet.
	ctx = xcontext.WithRequestUserID(ctx, testutil.Follower1.UserID)
	badges, err := domain.GetMyBadges(ctx, &model.GetMyBadgesRequest{
		CommunityID: testutil.Follower1.CommunityID,
	})
	require.NoError(t, err)
	require.Len(t, badges.Badges, 1)
	require.Equal(t, badge.SharpScoutBadgeName, badges.Badges[0].Name)
	require.Equal(t, 1, badges.Badges[0].Level)
	require.Equal(t, false, badges.Badges[0].WasNotified)

	// Get badges again and ensure they was notified to client.
	badges, err = domain.GetMyBadges(ctx, &model.GetMyBadgesRequest{
		CommunityID: testutil.Follower1.CommunityID,
	})
	require.NoError(t, err)
	require.Len(t, badges.Badges, 1)
	require.Equal(t, true, badges.Badges[0].WasNotified)
}
