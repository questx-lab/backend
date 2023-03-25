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
	suite.createUser()
	suite.createProject()

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
			setup: func() {
				suite.createUser()
				suite.createProject()
				suite.createCollaborator(entity.CollaboratorRoleOwner)
			},
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
			wantErr: errorx.NewGeneric(errorx.ErrBadRequest, "project not found"),
		},
		{
			name: "err permission by other user",
			setup: func() {
				suite.createCollaborator(entity.CollaboratorRoleOwner)
			},
			args: args{
				ctx: testutil.NewMockContextWithUserID("invalid-user"),
				req: &model.CreateCategoryRequest{
					ProjectID: suite.Project.ID,
					Name:      "valid-project",
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrPermissionDenied, "project not found"),
		},
		{
			name: "err user does not have permission",
			setup: func() {
				suite.createUser()
				suite.createProject()
				suite.createCollaborator(entity.CollaboratorRoleReviewer)
			},
			args: args{
				ctx: testutil.NewMockContextWithUserID(suite.User.ID),
				req: &model.CreateCategoryRequest{
					ProjectID: suite.Project.ID,
					Name:      "valid-project",
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrPermissionDenied, "project not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := domain.NewCategoryDomain(
				categoryRepo,
				projectRepo,
				collaboratorRepo,
			)

			got, err := d.Create(tt.args.ctx, tt.args.req)
			if err != nil && err != tt.wantErr {
				t.Errorf("categoryDomain.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("categoryDomain.Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
