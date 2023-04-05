package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
)

func Test_collaboratorDomain_Create(t *testing.T) {
	type args struct {
		ctx xcontext.Context
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
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User1.ID),
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
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User1.ID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: testutil.Project1.ID,
					UserID:    testutil.User1.ID,
					Role:      string(entity.Reviewer),
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Can not assign by yourself"),
		},
		{
			name: "wrong collaborator role",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User1.ID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: testutil.Project1.ID,
					UserID:    testutil.User2.ID,
					Role:      "wrong-role",
				},
			},
			wantErr: errorx.New(errorx.BadRequest, "Invalid role"),
		},
		{
			name: "invalid user",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User1.ID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: testutil.Project1.ID,
					UserID:    "invalid-user",
					Role:      string(entity.Reviewer),
				},
			},
			wantErr: errorx.New(errorx.NotFound, "Not found user"),
		},
		{
			name: "invalid project",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User1.ID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: "invalid-project-id",
					UserID:    testutil.User2.ID,
					Role:      string(entity.Reviewer),
				},
			},
			wantErr: errorx.New(errorx.NotFound, "Not found project"),
		},
		{
			name: "err user not have permission",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User3.ID),
				req: &model.CreateCollaboratorRequest{
					ProjectID: testutil.Project1.ID,
					UserID:    testutil.User2.ID,
					Role:      string(entity.Reviewer),
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			collaboratorRepo := repository.NewCollaboratorRepository()
			d := &collaboratorDomain{
				userRepo:         repository.NewUserRepository(),
				projectRepo:      repository.NewProjectRepository(),
				collaboratorRepo: collaboratorRepo,
				roleVerifier:     common.NewProjectRoleVerifier(collaboratorRepo),
			}

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
