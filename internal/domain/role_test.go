package domain

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
)

func Test_roleDomain_CreateRole(t *testing.T) {
	roleRepo := repository.NewRoleRepository()
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{}, &testutil.MockRedisClient{})
	roleVerifier := common.NewCommunityRoleVerifier(
		repository.NewFollowerRoleRepository(),
		repository.NewRoleRepository(),
		repository.NewUserRepository(&testutil.MockRedisClient{}),
	)
	type args struct {
		ctx context.Context
		req *model.CreateRoleRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.CreateRoleResponse
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User4.ID),
				req: &model.CreateRoleRequest{
					CommunityHandle: testutil.Community1.Handle,
					Name:            "role-1",
					Permissions:     uint64(entity.MANAGE_ROLE),
				},
			},
		},
		{
			name: "empty community handle",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.CreateRoleRequest{
					Name:        "role-1",
					Permissions: uint64(entity.MANAGE_ROLE),
				},
			},
			wantErr: errorx.New(errorx.BadRequest, "Not allow empty community handle"),
		},

		{
			name: "permission denied",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User6.ID),
				req: &model.CreateRoleRequest{
					CommunityHandle: testutil.Community1.Handle,
					Name:            "role-1",
					Permissions:     uint64(entity.MANAGE_ROLE),
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &roleDomain{
				roleRepo:      roleRepo,
				communityRepo: communityRepo,
				roleVerifier:  roleVerifier,
			}
			testutil.CreateFixtureDb(tt.args.ctx)
			req := httptest.NewRequest("GET", "/createRole", nil)
			ctx := xcontext.WithHTTPRequest(tt.args.ctx, req)
			_, err := d.CreateRole(ctx, tt.args.req)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_roleDomain_UpdateRole(t *testing.T) {
	roleRepo := repository.NewRoleRepository()
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{}, &testutil.MockRedisClient{})
	roleVerifier := common.NewCommunityRoleVerifier(
		repository.NewFollowerRoleRepository(),
		repository.NewRoleRepository(),
		repository.NewUserRepository(&testutil.MockRedisClient{}),
	)

	type args struct {
		ctx context.Context
		req *model.UpdateRoleRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.UpdateRoleResponse
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User4.ID),
				req: &model.UpdateRoleRequest{
					RoleID:      testutil.Role6.ID,
					Name:        "role-2",
					Permissions: uint64(entity.DELETE_COMMUNITY),
					Priority:    3,
				},
			},
		},
		{
			name: "low priority permission denied",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User6.ID),
				req: &model.UpdateRoleRequest{
					RoleID:      testutil.Role5.ID,
					Name:        "role-2",
					Permissions: uint64(entity.DELETE_COMMUNITY),
					Priority:    3,
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
		{
			name: "no permission permission denied",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User6.ID),
				req: &model.UpdateRoleRequest{
					RoleID:      testutil.Role5.ID,
					Name:        "role-2",
					Permissions: uint64(entity.DELETE_COMMUNITY),
					Priority:    3,
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
		{
			name: "no role permission denied",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User2.ID),
				req: &model.UpdateRoleRequest{
					RoleID:      testutil.Role5.ID,
					Name:        "role-2",
					Permissions: uint64(entity.DELETE_COMMUNITY),
					Priority:    3,
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &roleDomain{
				roleRepo:      roleRepo,
				communityRepo: communityRepo,
				roleVerifier:  roleVerifier,
			}
			testutil.CreateFixtureDb(tt.args.ctx)
			req := httptest.NewRequest("GET", "/updateRole", nil)
			ctx := xcontext.WithHTTPRequest(tt.args.ctx, req)
			_, err := d.UpdateRole(ctx, tt.args.req)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
