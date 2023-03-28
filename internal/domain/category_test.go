package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/testutil"
)

func Test_categoryDomain_Create(t *testing.T) {
	db := testutil.CreateFixtureDb()
	// TODO: define repositories
	categoryRepo := repository.NewCategoryRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	collaboratorRepo := repository.NewCollaboratorRepository(db)

	// define args
	type args struct {
		ctx router.Context
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
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
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
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
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
				ctx: testutil.NewMockContextWithUserID("invalid-user"),
				req: &model.CreateCategoryRequest{
					ProjectID: testutil.Project1.ID,
					Name:      "valid-project",
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "User does not have permission"),
		},
		{
			name: "err user role does not have permission",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User3.ID),
				req: &model.CreateCategoryRequest{
					ProjectID: testutil.Project1.ID,
					Name:      "valid-project",
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "User role does not have permission"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewCategoryDomain(
				categoryRepo,
				projectRepo,
				collaboratorRepo,
			)

			got, err := d.Create(tt.args.ctx, tt.args.req)
			if err != nil {
				if tt.wantErr == nil {
					t.Errorf("categoryDomain.Create() error = %v, wantErr = %v", err, tt.wantErr)
				} else if tt.wantErr.Error() != err.Error() {
					t.Errorf("categoryDomain.Create() error = %v, wantErr = %v", err, tt.wantErr)
				}
				return
			}
			if !reflectutil.PartialEqual(tt.want, got) {
				t.Errorf("categoryDomain.Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
