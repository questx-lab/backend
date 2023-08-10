package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
)

func Test_userDomain_GetMe_GetUser(t *testing.T) {
	ctx := testutil.MockContext(t)
	testutil.CreateFixtureDb(ctx)

	domain := NewUserDomain(
		repository.NewUserRepository(testutil.RedisClient(ctx)),
		repository.NewOAuth2Repository(),
		repository.NewFollowerRepository(),
		repository.NewFollowerRoleRepository(),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(ctx)),
		repository.NewClaimedQuestRepository(),
		nil, nil, testutil.RedisClient(ctx),
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
	ctx := testutil.MockContext(t)
	testutil.CreateFixtureDb(ctx)

	domain := NewUserDomain(
		repository.NewUserRepository(testutil.RedisClient(ctx)),
		repository.NewOAuth2Repository(),
		repository.NewFollowerRepository(),
		repository.NewFollowerRoleRepository(),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(ctx)),
		repository.NewClaimedQuestRepository(),
		nil, nil, testutil.RedisClient(ctx),
	)

	inviteResp, err := domain.GetInvite(ctx, &model.GetInviteRequest{
		InviteCode: testutil.Follower1.InviteCode,
	})
	require.NoError(t, err)
	require.Equal(t, inviteResp.Community.Handle, testutil.Community1.Handle)
}
