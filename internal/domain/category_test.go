package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/questx-lab/backend/pkg/testutil"
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
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.CreateCategoryRequest{
					CommunityID: testutil.Community1.ID,
					Name:        "valid-community",
				},
			},
			want: &model.CreateCategoryResponse{},
		},
		{
			name: "invalid community id",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User2.ID),
				req: &model.CreateCategoryRequest{
					CommunityID: "invalid-community-id",
					Name:        "valid-community",
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
		{
			name: "err user does not have permission",
			args: args{
				ctx: testutil.MockContextWithUserID("invalid-user"),
				req: &model.CreateCategoryRequest{
					CommunityID: testutil.Community1.ID,
					Name:        "valid-community",
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
		{
			name: "err user role does not have permission",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User3.ID),
				req: &model.CreateCategoryRequest{
					CommunityID: testutil.Community1.ID,
					Name:        "valid-community",
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
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
				repository.NewCollaboratorRepository(),
				repository.NewUserRepository(),
			)

			got, err := d.Create(tt.args.ctx, tt.args.req)
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
