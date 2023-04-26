package domain

import (
	"testing"
	"time"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
)

func Test_questDomain_Create_Failed(t *testing.T) {
	type args struct {
		ctx xcontext.Context
		req *model.CreateQuestRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "no permission",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User2.ID),
				req: &model.CreateQuestRequest{
					ProjectID: testutil.Project1.ID,
					Title:     "new-quest",
				},
			},
			wantErr: "Permission denied",
		},
		{
			name: "invalid category",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Project1.CreatedBy),
				req: &model.CreateQuestRequest{
					ProjectID:      testutil.Project1.ID,
					Title:          "new-quest",
					Type:           "visit_link",
					Recurrence:     "once",
					ConditionOp:    "or",
					Categories:     []string{"invalid-category"},
					ValidationData: map[string]any{"link": "http://example.com"},
				},
			},
			wantErr: "Invalid category",
		},
		{
			name: "not found one of two category",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Project1.CreatedBy),
				req: &model.CreateQuestRequest{
					ProjectID:      testutil.Project1.ID,
					Title:          "new-quest",
					Type:           "visit_link",
					Recurrence:     "once",
					ConditionOp:    "or",
					Categories:     []string{"category1", "invalid-category"},
					ValidationData: map[string]any{"link": "http://example.com"},
				},
			},
			wantErr: "Invalid category",
		},
		{
			name: "not found category with incorrect project",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Project2.CreatedBy),
				req: &model.CreateQuestRequest{
					ProjectID:      testutil.Project2.ID,
					Title:          "new-quest",
					Type:           "visit_link",
					Recurrence:     "once",
					ConditionOp:    "or",
					Categories:     []string{"category1"},
					ValidationData: map[string]any{"link": "http://example.com"},
				},
			},
			wantErr: "Invalid category",
		},
		{
			name: "invalid validation data",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Project2.CreatedBy),
				req: &model.CreateQuestRequest{
					ProjectID:      testutil.Project2.ID,
					Title:          "new-quest",
					Type:           "visit_link",
					Recurrence:     "once",
					ConditionOp:    "or",
					Categories:     []string{"category1"},
					ValidationData: map[string]any{"link": "invalid url"},
				},
			},
			wantErr: "Invalid validation data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			questDomain := NewQuestDomain(
				repository.NewQuestRepository(),
				repository.NewProjectRepository(),
				repository.NewCategoryRepository(),
				repository.NewCollaboratorRepository(),
				repository.NewUserRepository(),
				nil,
			)

			_, err := questDomain.Create(tt.args.ctx, tt.args.req)
			require.Error(t, err)
			require.Equal(t, tt.wantErr, err.Error())
		})
	}
}

func Test_questDomain_Create_Successfully(t *testing.T) {
	ctx := testutil.NewMockContextWithUserID(nil, testutil.Project1.CreatedBy)
	testutil.CreateFixtureDb(ctx)
	questDomain := NewQuestDomain(
		repository.NewQuestRepository(),
		repository.NewProjectRepository(),
		repository.NewCategoryRepository(),
		repository.NewCollaboratorRepository(),
		repository.NewUserRepository(),
		nil,
	)

	createQuestReq := &model.CreateQuestRequest{
		ProjectID:      testutil.Project1.ID,
		Title:          "new-quest",
		Type:           "text",
		Recurrence:     "once",
		ConditionOp:    "or",
		Categories:     []string{"category1", "category2"},
		ValidationData: `{}`,
	}

	questResp, err := questDomain.Create(ctx, createQuestReq)
	require.NoError(t, err)
	require.NotEmpty(t, questResp.ID)

	var result entity.Quest
	tx := ctx.DB().Model(&entity.Quest{}).Take(&result, "id", questResp.ID)
	require.NoError(t, tx.Error)
	require.Equal(t, testutil.Project1.ID, result.ProjectID)
	require.Equal(t, entity.QuestDraft, result.Status)
	require.Equal(t, createQuestReq.Title, result.Title)
	require.Equal(t, entity.QuestText, result.Type)
	require.Equal(t, entity.Once, result.Recurrence)
	require.Equal(t, entity.Or, result.ConditionOp)
}

func Test_questDomain_Get(t *testing.T) {
	ctx := testutil.NewMockContextWithUserID(nil, testutil.Project1.CreatedBy)
	testutil.CreateFixtureDb(ctx)
	questDomain := NewQuestDomain(
		repository.NewQuestRepository(),
		repository.NewProjectRepository(),
		repository.NewCategoryRepository(),
		repository.NewCollaboratorRepository(),
		repository.NewUserRepository(),
		nil,
	)

	resp, err := questDomain.Get(ctx, &model.GetQuestRequest{ID: testutil.Quest1.ID})
	require.NoError(t, err)
	require.Equal(t, testutil.Quest1.Title, resp.Title)
	require.Equal(t, string(testutil.Quest1.Type), resp.Type)
	require.Equal(t, string(testutil.Quest1.Status), resp.Status)
	require.Equal(t, string(testutil.Quest1.Awards[0].Type), resp.Awards[0].Type)
	require.Equal(t, testutil.Quest1.Awards[0].Value, resp.Awards[0].Value)
	require.Equal(t, string(testutil.Quest1.Conditions[0].Type), resp.Conditions[0].Type)
	require.Equal(t, testutil.Quest1.Conditions[0].Op, resp.Conditions[0].Op)
	require.Equal(t, testutil.Quest1.Conditions[0].Value, resp.Conditions[0].Value)
	require.Equal(t, testutil.Quest1.CreatedAt.Format(time.RFC3339Nano), resp.CreatedAt)
	require.Equal(t, testutil.Quest1.UpdatedAt.Format(time.RFC3339Nano), resp.UpdatedAt)
}

func Test_questDomain_GetList(t *testing.T) {
	type args struct {
		ctx xcontext.Context
		req *model.GetListQuestRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.GetListQuestResponse
		wantErr bool
	}{
		{
			name: "get successfully",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Project1.CreatedBy),
				req: &model.GetListQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    0,
					Limit:     2,
				},
			},
			want: &model.GetListQuestResponse{
				Quests: []model.ShortQuest{
					{
						ID:         testutil.Quest1.ID,
						Type:       string(testutil.Quest1.Type),
						Title:      testutil.Quest1.Title,
						Status:     string(testutil.Quest1.Status),
						Categories: testutil.Quest1.CategoryIDs,
						Recurrence: string(testutil.Quest1.Recurrence),
					},
					{
						ID:         testutil.Quest2.ID,
						Type:       string(testutil.Quest2.Type),
						Title:      testutil.Quest2.Title,
						Status:     string(testutil.Quest2.Status),
						Categories: []string{},
						Recurrence: string(testutil.Quest2.Recurrence),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nagative limit",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Project1.CreatedBy),
				req: &model.GetListQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    0,
					Limit:     -1,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "exceed maximum limit",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Project1.CreatedBy),
				req: &model.GetListQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    0,
					Limit:     51,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			d := &questDomain{
				questRepo:    repository.NewQuestRepository(),
				projectRepo:  repository.NewProjectRepository(),
				roleVerifier: common.NewProjectRoleVerifier(repository.NewCollaboratorRepository(), repository.NewUserRepository()),
			}

			got, err := d.GetList(tt.args.ctx, tt.args.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// No need to check result if they are nil pointer.
			if tt.want == nil || got == nil {
				require.Equal(t, tt.want, got)
				return
			}

			require.Equal(t, len(tt.want.Quests), len(got.Quests))
			for i := range got.Quests {
				require.Equal(t, tt.want.Quests[i].ID, got.Quests[i].ID)
				require.Equal(t, tt.want.Quests[i].Type, got.Quests[i].Type)
				require.Equal(t, tt.want.Quests[i].Title, got.Quests[i].Title)
				require.Equal(t, tt.want.Quests[i].Status, got.Quests[i].Status)
				require.Equal(t, tt.want.Quests[i].Recurrence, got.Quests[i].Recurrence)
				require.Equal(t, len(tt.want.Quests[i].Categories), len(got.Quests[i].Categories))
				for j := range got.Quests[i].Categories {
					require.Equal(t, tt.want.Quests[i].Categories[j], got.Quests[i].Categories[j])
				}
			}
		})
	}
}
