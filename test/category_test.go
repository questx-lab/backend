package test

import (
	"reflect"
	"testing"

	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/testutil"
)

func Test_categoryDomain_Create(t *testing.T) {
	suite := NewSuite()

	// TODO: define repositories
	categoryRepo := repository.NewCategoryRepository(suite.db)
	projectRepo := repository.NewProjectRepository(suite.db)
	collaboratorRepo := repository.NewCollaboratorRepository(suite.db)

	// TODO: define steps
	_ = suite.createUser()
	_ = suite.createProject()
	_ = suite.createCollaborator(entity.CollaboratorRoleOwner)

	//* define args
	type args struct {
		ctx router.Context
		req *model.CreateCategoryRequest
	}

	tests := []struct {
		name    string
		args    args
		want    *model.CreateCategoryResponse
		wantErr error
		setup   func()
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.NewMockContextWithUserID(suite.User.ID),
				req: &model.CreateCategoryRequest{
					ProjectID: suite.Project.ID,
					Name:      "valid-project",
				},
			},
			want: &model.CreateCategoryResponse{
				Success: true,
			},
		},
		{
			name: "invalid project id",
			args: args{
				ctx: testutil.NewMockContextWithUserID(suite.User.ID),
				req: &model.CreateCategoryRequest{
					ProjectID: "invalid-project-id",
					Name:      "valid-project",
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrNotFound, "project not found"),
		},
		{
			name: "err user does not have permission",
			args: args{
				ctx: testutil.NewMockContextWithUserID("invalid-user"),
				req: &model.CreateCategoryRequest{
					ProjectID: suite.Project.ID,
					Name:      "valid-project",
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrPermissionDenied, "user does not have permission"),
		},
		{
			name: "err user role does not have permission",
			setup: func() {
				_ = suite.updateCollaboratorRole(entity.CollaboratorRoleReviewer)
			},
			args: args{
				ctx: testutil.NewMockContextWithUserID(suite.User.ID),
				req: &model.CreateCategoryRequest{
					ProjectID: suite.Project.ID,
					Name:      "valid-project",
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrPermissionDenied, "user role does not have permission"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := domain.NewCategoryDomain(
				categoryRepo,
				projectRepo,
				collaboratorRepo,
			)

			suite.db.Find(&entity.Collaborator{})
			if tt.setup != nil {
				tt.setup()
			}

			got, err := d.Create(tt.args.ctx, tt.args.req)
			if err != nil {
				if tt.wantErr == nil {
					t.Errorf("categoryDomain.Create() error = %v, wantErr = %v", err, tt.wantErr)
				} else if tt.wantErr.Error() != err.Error() {
					t.Errorf("categoryDomain.Create() error = %v, wantErr = %v", err, tt.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("categoryDomain.Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
