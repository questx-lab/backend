package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/structutil"
	"github.com/questx-lab/backend/pkg/testutil"
)

func Test_collaboratorDomain_Create(t *testing.T) {
	suite := NewSuite(t)

	// TODO: define repositories
	userRepo := repository.NewUserRepository(suite.db)
	projectRepo := repository.NewProjectRepository(suite.db)
	collaboratorRepo := repository.NewCollaboratorRepository(suite.db)

	// TODO: define steps
	// owner
	_ = suite.createUser()
	_ = suite.createProject()
	_ = suite.createCollaborator(entity.Owner)
	ownerID := suite.User.ID

	// reviewer
	_ = suite.createUser()
	_ = suite.createCollaborator(entity.Reviewer)
	reviewerID := suite.User.ID

	// new collaborator
	_ = suite.createUser()
	collaboratorID := suite.User.ID

	type args struct {
		ctx router.Context
		req *model.CreateCollaboratorRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.CreateCollaboratorResponse
		wantErr error
		setup   func()
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.NewMockContextWithUserID(ownerID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: suite.Project.ID,
					UserID:    collaboratorID,
					Role:      string(entity.Reviewer),
				},
			},
			want: &model.CreateCollaboratorResponse{},
		},
		{
			name: "err update by yourself",
			args: args{
				ctx: testutil.NewMockContextWithUserID(ownerID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: suite.Project.ID,
					UserID:    ownerID,
					Role:      string(entity.Reviewer),
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrPermissionDenied, "can not assign by yourself"),
		},
		{
			name: "wrong collaborator role",
			args: args{
				ctx: testutil.NewMockContextWithUserID(ownerID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: suite.Project.ID,
					UserID:    collaboratorID,
					Role:      "wrong-role",
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrBadRequest, "role is invalid"),
		},
		{
			name: "invalid user",
			args: args{
				ctx: testutil.NewMockContextWithUserID(ownerID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: suite.Project.ID,
					UserID:    "invalid-user",
					Role:      string(entity.Reviewer),
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrNotFound, "user not found"),
		},
		{
			name: "invalid project",
			args: args{
				ctx: testutil.NewMockContextWithUserID(ownerID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: "invalid-project-id",
					UserID:    collaboratorID,
					Role:      string(entity.Reviewer),
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrNotFound, "project not found"),
		},
		{
			name: "err user not have permission",
			args: args{
				ctx: testutil.NewMockContextWithUserID(reviewerID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: suite.Project.ID,
					UserID:    collaboratorID,
					Role:      string(entity.Reviewer),
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrPermissionDenied, "user role does not have permission"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &collaboratorDomain{
				projectRepo:      projectRepo,
				collaboratorRepo: collaboratorRepo,
				userRepo:         userRepo,
			}
			got, err := d.Create(tt.args.ctx, tt.args.req)
			if err != nil {
				if tt.wantErr == nil {
					t.Errorf("collaboratorDomain.Create() error = %v, want no error", err)
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("collaboratorDomain.Create() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else if !structutil.PartialEqual(tt.want, got) {
				t.Errorf("collaboratorDomain.Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
