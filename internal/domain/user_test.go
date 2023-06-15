package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func Test_userDomain_GetReferralInfo(t *testing.T) {
	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)

	domain := NewUserDomain(
		repository.NewUserRepository(),
		repository.NewOAuth2Repository(),
		repository.NewFollowerRepository(),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
		badge.NewManager(
			repository.NewBadgeRepository(),
			repository.NewBadgeDetailRepository(),
			&testutil.MockBadge{
				NameValue:     badge.SharpScoutBadgeName,
				IsGlobalValue: false,
				ScanFunc: func(ctx context.Context, userID, communityID string) ([]entity.Badge, error) {
					return nil, nil
				},
			},
		),
		nil,
	)

	inviteResp, err := domain.GetInvite(ctx, &model.GetInviteRequest{
		InviteCode: testutil.Follower1.InviteCode,
	})
	require.NoError(t, err)
	require.Equal(t, inviteResp.Community.Handle, testutil.Community1.Handle)
}
