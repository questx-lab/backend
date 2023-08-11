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

func Test_categoryDomain_Create(t *testing.T) {
	// define args
	type args struct {
		ctx context.Context
		req *model.CreateCategoryRequest
	}

	tests := []struct {
		name    string
		args    args
		want    *model.CreateCategoryResponse
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.MockContextWithUserID(t, testutil.User1.ID),
				req: &model.CreateCategoryRequest{
					CommunityHandle: testutil.Community1.Handle,
					Name:            "valid-community",
				},
			},
			want: &model.CreateCategoryResponse{},
		},
		{
			name: "invalid community id",
			args: args{
				ctx: testutil.MockContextWithUserID(t, testutil.User2.ID),
				req: &model.CreateCategoryRequest{
					CommunityHandle: "invalid-community-id",
					Name:            "valid-community",
				},
			},
			wantErr: errorx.New(errorx.NotFound, "Not found community"),
		},
		{
			name: "err user does not have permission",
			args: args{
				ctx: testutil.MockContextWithUserID(t, "invalid-user"),
				req: &model.CreateCategoryRequest{
					CommunityHandle: testutil.Community1.Handle,
					Name:            "valid-community",
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
		{
			name: "err user role does not have permission",
			args: args{
				ctx: testutil.MockContextWithUserID(t, testutil.User2.ID),
				req: &model.CreateCategoryRequest{
					CommunityHandle: testutil.Community1.Handle,
					Name:            "valid-community",
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			d := NewCategoryDomain(
				repository.NewCategoryRepository(),
				repository.NewQuestRepository(nil),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(tt.args.ctx)),
				common.NewCommunityRoleVerifier(
					repository.NewFollowerRoleRepository(),
					repository.NewRoleRepository(),
					repository.NewUserRepository(testutil.RedisClient(tt.args.ctx)),
				),
			)
			req := httptest.NewRequest("GET", "/createCategory", nil)
			ctx := xcontext.WithHTTPRequest(tt.args.ctx, req)
			got, err := d.Create(ctx, tt.args.req)
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
