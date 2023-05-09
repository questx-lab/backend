package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/reflectutil"
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
					Status:         "draft",
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
					Status:         "draft",
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
					Status:         "draft",
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
					Status:         "active",
					Recurrence:     "once",
					ConditionOp:    "or",
					Categories:     []string{"category1"},
					ValidationData: map[string]any{"link": "invalid url"},
				},
			},
			wantErr: "Invalid validation data",
		},
		{
			name: "invalid status",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Project2.CreatedBy),
				req: &model.CreateQuestRequest{
					ProjectID:      testutil.Project2.ID,
					Title:          "new-quest",
					Type:           "visit_link",
					Status:         "something",
					Recurrence:     "once",
					ConditionOp:    "or",
					Categories:     []string{"category1"},
					ValidationData: map[string]any{"link": "invalid url"},
				},
			},
			wantErr: "Invalid quest status something",
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
				repository.NewClaimedQuestRepository(),
				repository.NewOAuth2Repository(),
				nil, nil, nil,
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
		repository.NewClaimedQuestRepository(),
		repository.NewOAuth2Repository(),
		nil, nil, nil,
	)

	createQuestReq := &model.CreateQuestRequest{
		ProjectID:      testutil.Project1.ID,
		Title:          "new-quest",
		Type:           "text",
		Status:         "active",
		Recurrence:     "once",
		ConditionOp:    "or",
		Categories:     []string{"category1", "category2"},
		ValidationData: map[string]any{},
	}

	questResp, err := questDomain.Create(ctx, createQuestReq)
	require.NoError(t, err)
	require.NotEmpty(t, questResp.ID)

	var result entity.Quest
	tx := ctx.DB().Model(&entity.Quest{}).Take(&result, "id", questResp.ID)
	require.NoError(t, tx.Error)
	require.Equal(t, testutil.Project1.ID, result.ProjectID)
	require.Equal(t, createQuestReq.Status, string(result.Status))
	require.Equal(t, createQuestReq.Title, result.Title)
	require.Equal(t, entity.QuestText, result.Type)
	require.Equal(t, entity.Once, result.Recurrence)
	require.Equal(t, entity.Or, result.ConditionOp)
}

func Test_questDomain_Get(t *testing.T) {
	type args struct {
		ctx xcontext.Context
		req *model.GetQuestRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.GetQuestResponse
		wantErr bool
	}{
		{
			name: "get successfully",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Project1.CreatedBy),
				req: &model.GetQuestRequest{
					ID: testutil.Quest1.ID,
				},
			},
			want: &model.GetQuestResponse{
				ID:         testutil.Quest1.ID,
				Type:       string(testutil.Quest1.Type),
				Title:      testutil.Quest1.Title,
				Status:     string(testutil.Quest1.Status),
				Categories: testutil.Quest1.CategoryIDs,
				Recurrence: string(testutil.Quest1.Recurrence),
			},
			wantErr: false,
		},
		{
			name: "include not claimable reason",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User3.ID),
				req: &model.GetQuestRequest{
					ID:                       testutil.Quest2.ID,
					IncludeUnclaimableReason: true,
				},
			},
			want: &model.GetQuestResponse{

				ID:                testutil.Quest2.ID,
				Type:              string(testutil.Quest2.Type),
				Title:             testutil.Quest2.Title,
				Status:            string(testutil.Quest2.Status),
				Categories:        testutil.Quest2.CategoryIDs,
				Recurrence:        string(testutil.Quest2.Recurrence),
				UnclaimableReason: "Please complete quest Quest 1 before claiming this quest",
			},
			wantErr: false,
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
				repository.NewClaimedQuestRepository(),
				repository.NewOAuth2Repository(),
				nil, nil, nil,
			)

			got, err := questDomain.Get(tt.args.ctx, tt.args.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.want != nil {
				require.True(t, reflectutil.PartialEqual(tt.want, got), "%v != %v", tt.want, got)
			}
		})
	}
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
				Quests: []model.Quest{
					{
						ID:         testutil.Quest3.ID,
						Type:       string(testutil.Quest3.Type),
						Title:      testutil.Quest3.Title,
						Status:     string(testutil.Quest3.Status),
						Categories: testutil.Quest3.CategoryIDs,
						Recurrence: string(testutil.Quest3.Recurrence),
					},
					{
						ID:         testutil.Quest2.ID,
						Type:       string(testutil.Quest2.Type),
						Title:      testutil.Quest2.Title,
						Status:     string(testutil.Quest2.Status),
						Categories: testutil.Quest2.CategoryIDs,
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
			want: &model.GetListQuestResponse{
				Quests: []model.Quest{
					{
						ID:         testutil.Quest3.ID,
						Type:       string(testutil.Quest3.Type),
						Title:      testutil.Quest3.Title,
						Status:     string(testutil.Quest3.Status),
						Categories: testutil.Quest3.CategoryIDs,
						Recurrence: string(testutil.Quest3.Recurrence),
					},
					{
						ID:         testutil.Quest2.ID,
						Type:       string(testutil.Quest2.Type),
						Title:      testutil.Quest2.Title,
						Status:     string(testutil.Quest2.Status),
						Categories: testutil.Quest2.CategoryIDs,
						Recurrence: string(testutil.Quest2.Recurrence),
					},
					{
						ID:         testutil.Quest1.ID,
						Type:       string(testutil.Quest1.Type),
						Title:      testutil.Quest1.Title,
						Status:     string(testutil.Quest1.Status),
						Categories: testutil.Quest1.CategoryIDs,
						Recurrence: string(testutil.Quest1.Recurrence),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "include not claimable reason",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User3.ID),
				req: &model.GetListQuestRequest{
					ProjectID:                testutil.Project1.ID,
					Offset:                   0,
					Limit:                    2,
					IncludeUnclaimableReason: true,
				},
			},
			want: &model.GetListQuestResponse{
				Quests: []model.Quest{
					{
						ID:         testutil.Quest3.ID,
						Type:       string(testutil.Quest3.Type),
						Title:      testutil.Quest3.Title,
						Status:     string(testutil.Quest3.Status),
						Categories: testutil.Quest3.CategoryIDs,
						Recurrence: string(testutil.Quest3.Recurrence),
					},
					{
						ID:                testutil.Quest2.ID,
						Type:              string(testutil.Quest2.Type),
						Title:             testutil.Quest2.Title,
						Status:            string(testutil.Quest2.Status),
						Categories:        testutil.Quest2.CategoryIDs,
						Recurrence:        string(testutil.Quest2.Recurrence),
						UnclaimableReason: "Please complete quest Quest 1 before claiming this quest",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			d := NewQuestDomain(
				repository.NewQuestRepository(),
				repository.NewProjectRepository(),
				repository.NewCategoryRepository(),
				repository.NewCollaboratorRepository(),
				repository.NewUserRepository(),
				repository.NewClaimedQuestRepository(),
				repository.NewOAuth2Repository(),
				&testutil.MockTwitterEndpoint{},
				&testutil.MockDiscordEndpoint{},
				nil,
			)

			got, err := d.GetList(tt.args.ctx, tt.args.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.want != nil {
				require.True(t, reflectutil.PartialEqual(tt.want, got), "%v != %v", tt.want, got)
			}
		})
	}
}
