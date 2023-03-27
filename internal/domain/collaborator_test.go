package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/testutil"
)

func Test_collaboratorDomain_Create(t *testing.T) {
	db := testutil.CreateFixtureDb()
	// TODO: define repositories
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	collaboratorRepo := repository.NewCollaboratorRepository(db)

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
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: testutil.Project1.ID,
					UserID:    testutil.User2.ID,
					Role:      string(entity.Reviewer),
				},
			},
			want: &model.CreateCollaboratorResponse{},
		},
		{
			name: "err update by yourself",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: testutil.Project1.ID,
					UserID:    testutil.User1.ID,
					Role:      string(entity.Reviewer),
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrPermissionDenied, "Can not assign by yourself"),
		},
		{
			name: "wrong collaborator role",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: testutil.Project1.ID,
					UserID:    testutil.User2.ID,
					Role:      "wrong-role",
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrBadRequest, "Role is invalid"),
		},
		{
			name: "invalid user",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: testutil.Project1.ID,
					UserID:    "invalid-user",
					Role:      string(entity.Reviewer),
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrNotFound, "User not found"),
		},
		{
			name: "invalid project",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: "invalid-project-id",
					UserID:    testutil.User2.ID,
					Role:      string(entity.Reviewer),
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrNotFound, "Project not found"),
		},
		{
			name: "err user not have permission",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User3.ID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: testutil.Project1.ID,
					UserID:    testutil.User2.ID,
					Role:      string(entity.Reviewer),
				},
			},
			wantErr: errorx.NewGeneric(errorx.ErrPermissionDenied, "User role does not have permission"),
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
			} else if !reflectutil.PartialEqual(tt.want, got) {
				t.Errorf("collaboratorDomain.Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
