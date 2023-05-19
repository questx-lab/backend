package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func Test_collaboratorDomain_Assign(t *testing.T) {
	type args struct {
		ctx context.Context
		req *model.AssignCollaboratorRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.AssignCollaboratorResponse
		wantErr error
		setup   func()
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.AssignCollaboratorRequest{
					ProjectID: testutil.Project1.ID,
					UserID:    testutil.User2.ID,
					Role:      string(entity.Reviewer),
				},
			},
			want: &model.AssignCollaboratorResponse{},
		},
		{
			name: "err update by yourself",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.AssignCollaboratorRequest{
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
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.AssignCollaboratorRequest{
					ProjectID: testutil.Project1.ID,
					UserID:    testutil.User2.ID,
					Role:      "wrong-role",
				},
			},
			wantErr: errorx.New(errorx.BadRequest, "Invalid role"),
		},
		{
			name: "err user not have permission",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User3.ID),
				req: &model.AssignCollaboratorRequest{
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
			userRepo := repository.NewUserRepository()
			d := &collaboratorDomain{
				userRepo:         repository.NewUserRepository(),
				projectRepo:      repository.NewProjectRepository(),
				collaboratorRepo: collaboratorRepo,
				roleVerifier:     common.NewProjectRoleVerifier(collaboratorRepo, userRepo),
			}

			got, err := d.Assign(tt.args.ctx, tt.args.req)
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

func Test_projectDomain_GetMyCollabs(t *testing.T) {
	ctx := testutil.MockContextWithUserID(testutil.Project1.CreatedBy)
	testutil.CreateFixtureDb(ctx)
	projectRepo := repository.NewProjectRepository()
	collaboratorRepo := repository.NewCollaboratorRepository()
	userRepo := repository.NewUserRepository()
	domain := NewCollaboratorDomain(projectRepo, collaboratorRepo, userRepo)
	result, err := domain.GetMyCollabs(ctx, &model.GetMyCollabsRequest{
		Offset: 0,
		Limit:  10,
	})

	require.NoError(t, err)
	require.Equal(t, 1, len(result.Collaborators))

	actual := result.Collaborators[0]

	expected := model.Collaborator{
		UserID:    testutil.Collaborator1.UserID,
		ProjectID: testutil.Collaborator1.ProjectID,
		Project: model.Project{
			ID:           testutil.Project1.ID,
			CreatedBy:    testutil.Project1.CreatedBy,
			Introduction: string(testutil.Project1.Introduction),
			Name:         testutil.Project1.Name,
			Twitter:      testutil.Project1.Twitter,
			Discord:      testutil.Project1.Discord,
		},
		Role:      string(testutil.Collaborator1.Role),
		CreatedBy: testutil.Collaborator1.CreatedBy,
	}

	require.True(t, reflectutil.PartialEqual(expected, actual), "%v != %v", expected, actual)
}
