package domain

import (
	"testing"

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
		ctx xcontext.Context
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
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User1.ID),
				req: &model.CreateCategoryRequest{
					ProjectID: testutil.Project1.ID,
					Name:      "valid-project",
				},
			},
			want: &model.CreateCategoryResponse{},
		},
		{
			name: "invalid project id",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User1.ID),
				req: &model.CreateCategoryRequest{
					ProjectID: "invalid-project-id",
					Name:      "valid-project",
				},
			},
			wantErr: errorx.New(errorx.NotFound, "Not found project"),
		},
		{
			name: "err user does not have permission",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, "invalid-user"),
				req: &model.CreateCategoryRequest{
					ProjectID: testutil.Project1.ID,
					Name:      "valid-project",
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
		{
			name: "err user role does not have permission",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User3.ID),
				req: &model.CreateCategoryRequest{
					ProjectID: testutil.Project1.ID,
					Name:      "valid-project",
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
				repository.NewProjectRepository(),
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
