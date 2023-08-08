package domain

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
)

func Test_communityDomain_TransferCommunity(t *testing.T) {
	ctx := testutil.MockContextWithUserID(testutil.User1.ID)
	testutil.CreateFixtureDb(ctx)
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{}, &testutil.MockRedisClient{})
	roleRepo := repository.NewRoleRepository()
	followerRepo := repository.NewFollowerRepository()
	followerRoleRepo := repository.NewFollowerRoleRepository()
	userRepo := repository.NewUserRepository(&testutil.MockRedisClient{})
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	oauth2Repo := repository.NewOAuth2Repository()
	chatChannelRepo := repository.NewChatChannelRepository()
	domain := NewCommunityDomain(
		communityRepo, followerRepo, followerRoleRepo, userRepo, questRepo,
		oauth2Repo, chatChannelRepo, roleRepo, nil, nil, nil, nil,
		common.NewCommunityRoleVerifier(
			repository.NewFollowerRoleRepository(),
			repository.NewRoleRepository(),
			repository.NewUserRepository(&testutil.MockRedisClient{}),
		),
		&testutil.MockRedisClient{},
	)
	type args struct {
		ctx context.Context
		req *model.TransferCommunityRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.TransferCommunityResponse
		wantErr error
		setup   func()
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.TransferCommunityRequest{
					CommunityHandle: testutil.Community2.Handle,
					ToUserID:        testutil.User3.ID,
				},
			},
			want: &model.TransferCommunityResponse{},
		},
		{
			name: "err user not found",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.TransferCommunityRequest{
					CommunityHandle: testutil.Community2.Handle,
					ToUserID:        "wrong_to_id",
				},
			},
			wantErr: errorx.New(errorx.NotFound, "Not found user"),
		},
		{
			name: "err community not found",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.TransferCommunityRequest{
					CommunityHandle: "community not found",
					ToUserID:        testutil.User2.ID,
				},
			},
			wantErr: errorx.New(errorx.NotFound, "Not found community"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)

			got, err := domain.TransferCommunity(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tt.wantErr.Error(), err.Error())
			}

			if tt.want == nil {
				require.Nil(t, got)
			} else {
				require.True(t, reflectutil.PartialEqual(tt.want, got), "%v != %v", tt.want, got)
			}
		})
	}

}

func Test_communityDomain_TransferCommunity_multi_transfer(t *testing.T) {
	ctx := testutil.MockContextWithUserID(testutil.User1.ID)
	testutil.CreateFixtureDb(ctx)
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{}, &testutil.MockRedisClient{})
	roleRepo := repository.NewRoleRepository()
	followerRepo := repository.NewFollowerRepository()
	followerRoleRepo := repository.NewFollowerRoleRepository()
	userRepo := repository.NewUserRepository(&testutil.MockRedisClient{})
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	oauth2Repo := repository.NewOAuth2Repository()
	chatChannelRepo := repository.NewChatChannelRepository()
	domain := NewCommunityDomain(
		communityRepo, followerRepo, followerRoleRepo, userRepo, questRepo,
		oauth2Repo, chatChannelRepo, roleRepo, nil, nil, nil, nil,
		common.NewCommunityRoleVerifier(
			repository.NewFollowerRoleRepository(),
			repository.NewRoleRepository(),
			repository.NewUserRepository(&testutil.MockRedisClient{}),
		),
		&testutil.MockRedisClient{},
	)

	req := &model.TransferCommunityRequest{
		CommunityHandle: testutil.Community2.Handle,
		ToUserID:        testutil.User3.ID,
	}

	_, err := domain.TransferCommunity(ctx, req)
	require.NoError(t, err)

	req = &model.TransferCommunityRequest{
		CommunityHandle: testutil.Community2.Handle,
		ToUserID:        testutil.User2.ID,
	}

	_, err = domain.TransferCommunity(ctx, req)
	require.NoError(t, err)
}

func Test_communityDomain_AssignRole(t *testing.T) {
	ctx := testutil.MockContextWithUserID(testutil.User1.ID)
	testutil.CreateFixtureDb(ctx)
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{}, &testutil.MockRedisClient{})
	roleRepo := repository.NewRoleRepository()
	followerRepo := repository.NewFollowerRepository()
	followerRoleRepo := repository.NewFollowerRoleRepository()
	userRepo := repository.NewUserRepository(&testutil.MockRedisClient{})
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	oauth2Repo := repository.NewOAuth2Repository()
	chatChannelRepo := repository.NewChatChannelRepository()
	domain := NewCommunityDomain(
		communityRepo, followerRepo, followerRoleRepo, userRepo, questRepo,
		oauth2Repo, chatChannelRepo, roleRepo, nil, nil, nil, nil,
		common.NewCommunityRoleVerifier(
			repository.NewFollowerRoleRepository(),
			repository.NewRoleRepository(),
			repository.NewUserRepository(&testutil.MockRedisClient{}),
		),
		&testutil.MockRedisClient{},
	)
	type args struct {
		ctx context.Context
		req *model.AssignRoleRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.AssignRoleResponse
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User6.ID),
				req: &model.AssignRoleRequest{
					UserID: testutil.User5.ID,
					RoleID: testutil.Role6.ID,
				},
			},
		},
		{
			name: "permission denied",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User5.ID),
				req: &model.AssignRoleRequest{
					UserID: testutil.User6.ID,
					RoleID: testutil.Role5.ID,
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			req := httptest.NewRequest("GET", "/assignCommunityRole", nil)
			ctx := xcontext.WithHTTPRequest(tt.args.ctx, req)
			_, err := domain.AssignRole(ctx, tt.args.req)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
