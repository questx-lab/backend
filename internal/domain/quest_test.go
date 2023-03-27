package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func Test_questDomain_Create(t *testing.T) {
	db := testutil.DefaultTestDb(t)
	projectRepo := repository.NewProjectRepository(db)
	questRepo := repository.NewQuestRepository(db)

	questDomain := NewQuestDomain(questRepo, projectRepo)

	t.Run("create quest successfully", func(t *testing.T) {
		createQuestReq := &model.CreateQuestRequest{
			ProjectID:   testutil.Project1.ID,
			Title:       "new-quest",
			Type:        "Visit Link",
			Recurrence:  "Once",
			ConditionOp: "OR",
		}

		ctx := testutil.NewMockContextWithUserID(testutil.Project1.CreatedBy)
		questResp, err := questDomain.Create(ctx, createQuestReq)
		require.NoError(t, err)
		require.NotEmpty(t, questResp.ID)

		var result entity.Quest
		tx := db.Model(&entity.Quest{}).Take(&result, "id", questResp.ID)
		require.NoError(t, tx.Error)
		require.Equal(t, testutil.Project1.ID, result.ProjectID)
		require.Equal(t, entity.QuestStatusDraft, result.Status)
		require.Equal(t, createQuestReq.Title, result.Title)
		require.Equal(t, entity.QuestVisitLink, result.Type)
		require.Equal(t, entity.QuestRecurrenceOnce, result.Recurrence)
		require.Equal(t, entity.QuestConditionOpOr, result.ConditionOp)
	})

	t.Run("no perrmission to create quest", func(t *testing.T) {
		// 1. New project created by user 2
		otherProject := &entity.Project{
			Base: entity.Base{
				ID: "user2_project1",
			},
			Name:      "User2 Project1",
			CreatedBy: testutil.User2.ID,
		}
		err := projectRepo.Create(context.Background(), otherProject)
		require.NoError(t, err)

		// 2. Verify that user1 cannot create this project.
		createQuestReq := &model.CreateQuestRequest{
			ProjectID: otherProject.ID,
			Title:     "new-quest",
		}

		ctx := testutil.NewMockContextWithUserID(testutil.User1.ID)
		_, err = questDomain.Create(ctx, createQuestReq)
		genericError := errorx.Error{}
		require.ErrorAs(t, err, &genericError)
		require.Equal(t, errorx.ErrGeneric.Code, genericError.Code)
		require.Equal(t, "Permission denied", genericError.Message)
	})
}

func Test_questDomain_Get(t *testing.T) {
	db := testutil.CreateFixtureDb()
	projectRepo := repository.NewProjectRepository(db)
	questRepo := repository.NewQuestRepository(db)

	questDomain := NewQuestDomain(questRepo, projectRepo)

	ctx := testutil.NewMockContextWithUserID(testutil.Project1.CreatedBy)
	resp, err := questDomain.Get(ctx, &model.GetQuestRequest{ID: testutil.Quest1.ID})
	require.NoError(t, err)
	require.Equal(t, testutil.Quest1.Title, resp.Title)
	require.Equal(t, enum.ToString(testutil.Quest1.Type), resp.Type)
	require.Equal(t, enum.ToString(testutil.Quest1.Status), resp.Status)
	require.Equal(t, testutil.Quest1.Awards[0].Type, resp.Awards[0].Type)
	require.Equal(t, testutil.Quest1.Awards[0].Value, resp.Awards[0].Value)
	require.Equal(t, testutil.Quest1.Conditions[0].Type, resp.Conditions[0].Type)
	require.Equal(t, testutil.Quest1.Conditions[0].Op, resp.Conditions[0].Op)
	require.Equal(t, testutil.Quest1.Conditions[0].Value, resp.Conditions[0].Value)
}

func Test_questDomain_GetList(t *testing.T) {
	db := testutil.CreateFixtureDb()

	type fields struct {
		questRepo   repository.QuestRepository
		projectRepo repository.ProjectRepository
	}
	type args struct {
		ctx router.Context
		req *model.GetListQuestRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetListQuestResponse
		wantErr bool
	}{
		{
			name: "get successfully",
			fields: fields{
				questRepo:   repository.NewQuestRepository(db),
				projectRepo: repository.NewProjectRepository(db),
			},
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.Project1.CreatedBy),
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
						Type:       enum.ToString(testutil.Quest1.Type),
						Title:      testutil.Quest1.Title,
						Status:     enum.ToString(testutil.Quest1.Status),
						Categories: testutil.Quest1.CategoryIDs,
						Recurrence: enum.ToString(testutil.Quest1.Recurrence),
					},
					{
						ID:         testutil.Quest2.ID,
						Type:       enum.ToString(testutil.Quest2.Type),
						Title:      testutil.Quest2.Title,
						Status:     enum.ToString(testutil.Quest2.Status),
						Categories: []string{},
						Recurrence: enum.ToString(testutil.Quest2.Recurrence),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nagative limit",
			fields: fields{
				questRepo:   repository.NewQuestRepository(db),
				projectRepo: repository.NewProjectRepository(db),
			},
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.Project1.CreatedBy),
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
			fields: fields{
				questRepo:   repository.NewQuestRepository(db),
				projectRepo: repository.NewProjectRepository(db),
			},
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.Project1.CreatedBy),
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
			d := &questDomain{
				questRepo:   tt.fields.questRepo,
				projectRepo: tt.fields.projectRepo,
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
