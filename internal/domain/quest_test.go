package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
)

func Test_questDomain_Create_Failed(t *testing.T) {
	type args struct {
		ctx context.Context
		req *model.CreateQuestRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "no permission",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User2.ID),
				req: &model.CreateQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Title:           "new-quest",
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
		{
			name: "invalid category",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Community1.CreatedBy),
				req: &model.CreateQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Title:           "new-quest",
					Type:            "visit_link",
					Status:          "draft",
					Recurrence:      "once",
					ConditionOp:     "or",
					CategoryID:      "invalid-category",
					ValidationData:  map[string]any{"link": "http://example.com"},
				},
			},
			wantErr: errorx.New(errorx.NotFound, "Invalid category"),
		},
		{
			name: "not found category with incorrect community",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Community2.CreatedBy),
				req: &model.CreateQuestRequest{
					CommunityHandle: testutil.Community2.Handle,
					Title:           "new-quest",
					Type:            "visit_link",
					Status:          "draft",
					Recurrence:      "once",
					ConditionOp:     "or",
					CategoryID:      "category1",
					ValidationData:  map[string]any{"link": "http://example.com"},
				},
			},
			wantErr: errorx.New(errorx.BadRequest, "Category doesn't belong to community"),
		},
		{
			name: "invalid validation data",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Community2.CreatedBy),
				req: &model.CreateQuestRequest{
					CommunityHandle: testutil.Community2.Handle,
					Title:           "new-quest",
					Type:            "visit_link",
					Status:          "active",
					Recurrence:      "once",
					ConditionOp:     "or",
					CategoryID:      "category1",
					ValidationData:  map[string]any{"link": "invalid url"},
				},
			},
			wantErr: errorx.New(errorx.BadRequest, "Invalid link"),
		},
		{
			name: "invalid status",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Community2.CreatedBy),
				req: &model.CreateQuestRequest{
					CommunityHandle: testutil.Community2.Handle,
					Title:           "new-quest",
					Type:            "visit_link",
					Status:          "something",
					Recurrence:      "once",
					ConditionOp:     "or",
					CategoryID:      "category1",
					ValidationData:  map[string]any{"link": "invalid url"},
				},
			},
			wantErr: errorx.New(errorx.BadRequest, "Invalid quest status something"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			questDomain := NewQuestDomain(
				repository.NewQuestRepository(&testutil.MockSearchCaller{}),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
				repository.NewCategoryRepository(),
				repository.NewCollaboratorRepository(),
				repository.NewUserRepository(),
				repository.NewClaimedQuestRepository(),
				repository.NewOAuth2Repository(),
				repository.NewPayRewardRepository(),
				repository.NewFollowerRepository(),
				nil, nil, nil, &testutil.MockLeaderboard{},
			)

			_, err := questDomain.Create(tt.args.ctx, tt.args.req)
			require.Error(t, err)
			require.Equal(t, tt.wantErr, err)
		})
	}
}

func Test_questDomain_Create_Successfully(t *testing.T) {
	ctx := testutil.MockContextWithUserID(testutil.Community1.CreatedBy)
	testutil.CreateFixtureDb(ctx)
	questDomain := NewQuestDomain(
		repository.NewQuestRepository(&testutil.MockSearchCaller{}),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
		repository.NewCategoryRepository(),
		repository.NewCollaboratorRepository(),
		repository.NewUserRepository(),
		repository.NewClaimedQuestRepository(),
		repository.NewOAuth2Repository(),
		repository.NewPayRewardRepository(),
		repository.NewFollowerRepository(),
		nil, nil, nil, &testutil.MockLeaderboard{},
	)

	createQuestReq := &model.CreateQuestRequest{
		CommunityHandle: testutil.Community1.Handle,
		Title:           "new-quest",
		Type:            "text",
		Status:          "active",
		Recurrence:      "once",
		ConditionOp:     "or",
		CategoryID:      "category1",
		ValidationData:  map[string]any{},
	}

	questResp, err := questDomain.Create(ctx, createQuestReq)
	require.NoError(t, err)
	require.NotEmpty(t, questResp.ID)

	var result entity.Quest
	tx := xcontext.DB(ctx).Model(&entity.Quest{}).Take(&result, "id", questResp.ID)
	require.NoError(t, tx.Error)
	require.Equal(t, testutil.Community1.ID, result.CommunityID.String)
	require.Equal(t, createQuestReq.Status, string(result.Status))
	require.Equal(t, createQuestReq.Title, result.Title)
	require.Equal(t, entity.QuestText, result.Type)
	require.Equal(t, entity.Once, result.Recurrence)
	require.Equal(t, entity.Or, result.ConditionOp)
}

func Test_questDomain_Get(t *testing.T) {
	type args struct {
		ctx context.Context
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
				ctx: testutil.MockContextWithUserID(testutil.Community1.CreatedBy),
				req: &model.GetQuestRequest{
					ID: testutil.Quest1.ID,
				},
			},
			want: &model.GetQuestResponse{
				ID:     testutil.Quest1.ID,
				Type:   string(testutil.Quest1.Type),
				Title:  testutil.Quest1.Title,
				Status: string(testutil.Quest1.Status),
				Category: model.Category{
					ID:   testutil.Category1.ID,
					Name: testutil.Category1.Name,
				},
				Recurrence: string(testutil.Quest1.Recurrence),
			},
			wantErr: false,
		},
		{
			name: "include not claimable reason",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User3.ID),
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
				Category:          model.Category{},
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
				repository.NewQuestRepository(&testutil.MockSearchCaller{}),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
				repository.NewCategoryRepository(),
				repository.NewCollaboratorRepository(),
				repository.NewUserRepository(),
				repository.NewClaimedQuestRepository(),
				repository.NewOAuth2Repository(),
				repository.NewPayRewardRepository(),
				repository.NewFollowerRepository(),
				nil, nil, nil, &testutil.MockLeaderboard{},
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
		ctx context.Context
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
				ctx: testutil.MockContextWithUserID(testutil.Community1.CreatedBy),
				req: &model.GetListQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Offset:          0,
					Limit:           2,
				},
			},
			want: &model.GetListQuestResponse{
				Quests: []model.Quest{
					{
						ID:         testutil.Quest3.ID,
						Type:       string(testutil.Quest3.Type),
						Title:      testutil.Quest3.Title,
						Status:     string(testutil.Quest3.Status),
						Recurrence: string(testutil.Quest3.Recurrence),
					},
					{
						ID:         testutil.Quest2.ID,
						Type:       string(testutil.Quest2.Type),
						Title:      testutil.Quest2.Title,
						Status:     string(testutil.Quest2.Status),
						Recurrence: string(testutil.Quest2.Recurrence),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nagative limit",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Community1.CreatedBy),
				req: &model.GetListQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Offset:          0,
					Limit:           -1,
				},
			},
			want: &model.GetListQuestResponse{
				Quests: []model.Quest{
					{
						ID:         testutil.Quest3.ID,
						Type:       string(testutil.Quest3.Type),
						Title:      testutil.Quest3.Title,
						Status:     string(testutil.Quest3.Status),
						Recurrence: string(testutil.Quest3.Recurrence),
					},
					{
						ID:         testutil.Quest2.ID,
						Type:       string(testutil.Quest2.Type),
						Title:      testutil.Quest2.Title,
						Status:     string(testutil.Quest2.Status),
						Recurrence: string(testutil.Quest2.Recurrence),
					},
					{
						ID:         testutil.Quest1.ID,
						Type:       string(testutil.Quest1.Type),
						Title:      testutil.Quest1.Title,
						Status:     string(testutil.Quest1.Status),
						Recurrence: string(testutil.Quest1.Recurrence),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "include not claimable reason",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User3.ID),
				req: &model.GetListQuestRequest{
					CommunityHandle:          testutil.Community1.Handle,
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
						Recurrence: string(testutil.Quest3.Recurrence),
					},
					{
						ID:                testutil.Quest2.ID,
						Type:              string(testutil.Quest2.Type),
						Title:             testutil.Quest2.Title,
						Status:            string(testutil.Quest2.Status),
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
				repository.NewQuestRepository(&testutil.MockSearchCaller{}),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
				repository.NewCategoryRepository(),
				repository.NewCollaboratorRepository(),
				repository.NewUserRepository(),
				repository.NewClaimedQuestRepository(),
				repository.NewOAuth2Repository(),
				repository.NewPayRewardRepository(),
				repository.NewFollowerRepository(),
				&testutil.MockTwitterEndpoint{},
				&testutil.MockDiscordEndpoint{},
				nil,
				&testutil.MockLeaderboard{},
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

func Test_questDomain_Update(t *testing.T) {
	type args struct {
		ctx context.Context
		req *model.UpdateQuestRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "no permission",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User2.ID),
				req: &model.UpdateQuestRequest{
					ID:    testutil.Quest1.ID,
					Title: "new-quest",
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
		{
			name: "invalid category",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Community1.CreatedBy),
				req: &model.UpdateQuestRequest{
					ID:             testutil.Quest1.ID,
					Status:         "active",
					Title:          "new-quest",
					Type:           "visit_link",
					Recurrence:     "once",
					ConditionOp:    "or",
					CategoryID:     "invalid-category",
					ValidationData: map[string]any{"link": "http://example.com"},
				},
			},
			wantErr: errorx.New(errorx.NotFound, "Invalid category"),
		},
		{
			name: "invalid validation data",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Community1.CreatedBy),
				req: &model.UpdateQuestRequest{
					ID:             testutil.Quest1.ID,
					Title:          "new-quest",
					Status:         "active",
					Type:           "visit_link",
					Recurrence:     "once",
					ConditionOp:    "or",
					CategoryID:     "category1",
					ValidationData: map[string]any{"link": "invalid url"},
				},
			},
			wantErr: errorx.New(errorx.BadRequest, "Invalid link"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			questDomain := NewQuestDomain(
				repository.NewQuestRepository(&testutil.MockSearchCaller{}),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
				repository.NewCategoryRepository(),
				repository.NewCollaboratorRepository(),
				repository.NewUserRepository(),
				repository.NewClaimedQuestRepository(),
				repository.NewOAuth2Repository(),
				repository.NewPayRewardRepository(),
				repository.NewFollowerRepository(),
				nil, nil, nil, &testutil.MockLeaderboard{},
			)

			_, err := questDomain.Update(tt.args.ctx, tt.args.req)
			require.Error(t, err)
			require.Equal(t, tt.wantErr, err)
		})
	}
}

func Test_questDomain_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		req *model.DeleteQuestRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "no permission",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User2.ID),
				req: &model.DeleteQuestRequest{
					ID: testutil.Quest1.ID,
				},
			},
			wantErr: "Permission denied",
		},
		{
			name: "happy case",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Community1.CreatedBy),
				req: &model.DeleteQuestRequest{
					ID: testutil.Quest1.ID,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			questDomain := NewQuestDomain(
				repository.NewQuestRepository(&testutil.MockSearchCaller{}),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
				repository.NewCategoryRepository(),
				repository.NewCollaboratorRepository(),
				repository.NewUserRepository(),
				repository.NewClaimedQuestRepository(),
				repository.NewOAuth2Repository(),
				repository.NewPayRewardRepository(),
				repository.NewFollowerRepository(),
				nil, nil, nil, &testutil.MockLeaderboard{},
			)

			_, err := questDomain.Delete(tt.args.ctx, tt.args.req)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_questDomain_GetTemplates(t *testing.T) {
	type args struct {
		ctx context.Context
		req *model.GetQuestTemplatesRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.GetQuestTemplatestResponse
		wantErr error
	}{
		{
			name: "get successfully",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Community1.CreatedBy),
				req: &model.GetQuestTemplatesRequest{
					Offset: 0,
					Limit:  2,
				},
			},
			want: &model.GetQuestTemplatestResponse{
				Templates: []model.Quest{
					{
						ID:         testutil.QuestTemplate.ID,
						Type:       string(testutil.QuestTemplate.Type),
						Title:      testutil.QuestTemplate.Title,
						Status:     string(testutil.QuestTemplate.Status),
						Recurrence: string(testutil.QuestTemplate.Recurrence),
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			d := NewQuestDomain(
				repository.NewQuestRepository(&testutil.MockSearchCaller{}),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
				repository.NewCategoryRepository(),
				repository.NewCollaboratorRepository(),
				repository.NewUserRepository(),
				repository.NewClaimedQuestRepository(),
				repository.NewOAuth2Repository(),
				repository.NewPayRewardRepository(),
				repository.NewFollowerRepository(),
				&testutil.MockTwitterEndpoint{},
				&testutil.MockDiscordEndpoint{},
				nil,
				&testutil.MockLeaderboard{},
			)

			got, err := d.GetTemplates(tt.args.ctx, tt.args.req)
			if tt.wantErr != nil {
				require.ErrorIs(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
			}

			if tt.want != nil {
				require.True(t, reflectutil.PartialEqual(tt.want, got), "%v != %v", tt.want, got)
			}
		})
	}
}

func Test_questDomain_ParseTemplate(t *testing.T) {
	ctx := testutil.MockContextWithUserID(testutil.Community1.CreatedBy)
	testutil.CreateFixtureDb(ctx)
	questDomain := NewQuestDomain(
		repository.NewQuestRepository(&testutil.MockSearchCaller{}),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
		repository.NewCategoryRepository(),
		repository.NewCollaboratorRepository(),
		repository.NewUserRepository(),
		repository.NewClaimedQuestRepository(),
		repository.NewOAuth2Repository(),
		repository.NewPayRewardRepository(),
		repository.NewFollowerRepository(),
		nil, nil, nil, &testutil.MockLeaderboard{},
	)

	resp, err := questDomain.ParseTemplate(ctx, &model.ParseQuestTemplatesRequest{
		TemplateID:      testutil.QuestTemplate.ID,
		CommunityHandle: testutil.Community1.Handle,
	})
	require.NoError(t, err)
	require.Equal(t, "Quest of User1 Community1", resp.Quest.Title)
	require.Equal(t, "Description is written by user1 for User1 Community1", resp.Quest.Description)
}

func Test_questDomain_Update_Point(t *testing.T) {
	ctx := testutil.MockContextWithUserID(testutil.Community1.CreatedBy)
	testutil.CreateFixtureDb(ctx)
	questDomain := NewQuestDomain(
		repository.NewQuestRepository(&testutil.MockSearchCaller{}),
		repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
		repository.NewCategoryRepository(),
		repository.NewCollaboratorRepository(),
		repository.NewUserRepository(),
		repository.NewClaimedQuestRepository(),
		repository.NewOAuth2Repository(),
		repository.NewPayRewardRepository(),
		repository.NewFollowerRepository(),
		nil, nil, nil, &testutil.MockLeaderboard{},
	)

	_, err := questDomain.Update(ctx, &model.UpdateQuestRequest{
		ID:             testutil.Quest1.ID,
		Status:         string(testutil.Quest1.Status),
		Title:          testutil.Quest1.Title,
		Type:           string(testutil.Quest1.Type),
		Recurrence:     string(testutil.Quest1.Recurrence),
		ConditionOp:    string(testutil.Quest1.ConditionOp),
		CategoryID:     testutil.Quest1.CategoryID.String,
		ValidationData: testutil.Quest1.ValidationData,
		Points:         testutil.Quest1.Points - 20,
	})
	require.NoError(t, err)

	follower, err := repository.NewFollowerRepository().Get(ctx, testutil.User1.ID, testutil.Community1.ID)
	require.NoError(t, err)
	require.Equal(t, testutil.Follower1.Points-20, follower.Points)
}
