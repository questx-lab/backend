package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
)

func Test_userDomain_GetMe_GetUser(t *testing.T) {
	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)

	domain := NewUserDomain(
		repository.NewUserRepository(&testutil.MockRedisClient{}),
		repository.NewOAuth2Repository(),
		repository.NewFollowerRepository(),
		repository.NewFollowerRoleRepository(),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}, &testutil.MockRedisClient{}),
		repository.NewClaimedQuestRepository(),
		badge.NewManager(
			repository.NewBadgeRepository(),
			repository.NewBadgeDetailRepository(),
		),
		nil, nil, &testutil.MockRedisClient{},
	)

	// User1 calls getMe.
	ctx = xcontext.WithRequestUserID(ctx, testutil.User1.ID)
	getMeResp, err := domain.GetMe(ctx, &model.GetMeRequest{})
	require.NoError(t, err)
	require.Equal(t, getMeResp.User, model.User{
		ShortUser: model.ShortUser{
			ID:        testutil.User1.ID,
			Name:      testutil.User1.Name,
			AvatarURL: testutil.User1.ProfilePicture,
		},
		WalletAddress:      testutil.User1.WalletAddress.String,
		Role:               string(testutil.User1.Role),
		Services:           map[string]string{},
		ReferralCode:       testutil.User1.ReferralCode,
		IsNewUser:          testutil.User1.IsNewUser,
		TotalCommunities:   2,
		TotalClaimedQuests: 2,
	})

	// User2 call getUser with parameter User1.ID
	ctx = xcontext.WithRequestUserID(ctx, testutil.User2.ID)
	getUserResp, err := domain.GetUser(ctx, &model.GetUserRequest{UserID: testutil.User1.ID})
	require.NoError(t, err)
	require.Equal(t, getUserResp.User, model.User{
		ShortUser: model.ShortUser{
			ID:        testutil.User1.ID,
			Name:      testutil.User1.Name,
			AvatarURL: testutil.User1.ProfilePicture,
		},
		WalletAddress:      "",
		Role:               "",
		Services:           map[string]string{},
		ReferralCode:       testutil.User1.ReferralCode,
		IsNewUser:          false,
		TotalCommunities:   2,
		TotalClaimedQuests: 2,
	})
}

func Test_userDomain_GetReferralInfo(t *testing.T) {
	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)

	domain := NewUserDomain(
		repository.NewUserRepository(&testutil.MockRedisClient{}),
		repository.NewOAuth2Repository(),
		repository.NewFollowerRepository(),
		repository.NewFollowerRoleRepository(),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}, &testutil.MockRedisClient{}),
		repository.NewClaimedQuestRepository(),
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
		nil, nil, &testutil.MockRedisClient{},
	)

	inviteResp, err := domain.GetInvite(ctx, &model.GetInviteRequest{
		InviteCode: testutil.Follower1.InviteCode,
	})
	require.NoError(t, err)
	require.Equal(t, inviteResp.Community.Handle, testutil.Community1.Handle)
}
