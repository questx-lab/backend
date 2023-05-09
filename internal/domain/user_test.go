package domain

import (
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
	ctx := testutil.NewMockContext()
	testutil.CreateFixtureDb(ctx)

	domain := NewUserDomain(
		repository.NewUserRepository(),
		repository.NewParticipantRepository(),
		repository.NewBadgeRepository(),
		badge.NewManager(
			repository.NewBadgeRepository(),
			&testutil.MockBadge{
				NameValue:     badge.SharpScoutBadgeName,
				IsGlobalValue: false,
				ScanFunc: func(ctx xcontext.Context, userID, projectID string) (int, error) {
					return 0, nil
				},
			},
		),
	)

	inviteResp, err := domain.GetInvite(ctx, &model.GetInviteRequest{
		InviteCode: testutil.Participant1.InviteCode,
	})
	require.NoError(t, err)
	require.Equal(t, inviteResp.Project.ID, testutil.Project1.ID)
	require.Equal(t, inviteResp.Project.Name, testutil.Project1.Name)
}

func Test_userDomain_FollowProject_and_GetBadges(t *testing.T) {
	ctx := testutil.NewMockContext()
	testutil.CreateFixtureDb(ctx)

	userRepo := repository.NewUserRepository()
	pariticipantRepo := repository.NewParticipantRepository()
	badgeRepo := repository.NewBadgeRepository()

	newUser := &entity.User{Base: entity.Base{ID: uuid.NewString()}}
	require.NoError(t, userRepo.Create(ctx, newUser))

	domain := NewUserDomain(
		userRepo,
		pariticipantRepo,
		badgeRepo,
		badge.NewManager(
			badgeRepo,
			&testutil.MockBadge{
				NameValue:     badge.SharpScoutBadgeName,
				IsGlobalValue: false,
				ScanFunc: func(ctx xcontext.Context, userID, projectID string) (int, error) {
					return 1, nil
				},
			},
		),
	)

	ctx = testutil.NewMockContextWithUserID(ctx, newUser.ID)
	_, err := domain.FollowProject(ctx, &model.FollowProjectRequest{
		ProjectID: testutil.Participant1.ProjectID,
		InvitedBy: testutil.Participant1.UserID,
	})
	require.NoError(t, err)

	// Get badges and check their level, name. Ensure that they haven't been
	// notified to client yet.
	badges, err := domain.GetBadges(ctx, &model.GetBadgesRequest{
		UserID:    testutil.Participant1.UserID,
		ProjectID: testutil.Participant1.ProjectID,
	})
	require.NoError(t, err)
	require.Len(t, badges.Badges, 1)
	require.Equal(t, badge.SharpScoutBadgeName, badges.Badges[0].Name)
	require.Equal(t, 1, badges.Badges[0].Level)
	require.Equal(t, false, badges.Badges[0].WasNotified)

	// Get badges again and ensure they was notified to client.
	badges, err = domain.GetBadges(ctx, &model.GetBadgesRequest{
		UserID:    testutil.Participant1.UserID,
		ProjectID: testutil.Participant1.ProjectID,
	})
	require.NoError(t, err)
	require.Len(t, badges.Badges, 1)
	require.Equal(t, true, badges.Badges[0].WasNotified)
}
