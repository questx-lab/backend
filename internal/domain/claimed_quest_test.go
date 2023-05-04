package domain

import (
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"

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

func Test_claimedQuestDomain_Claim_AutoText(t *testing.T) {
	ctx := testutil.NewMockContext()
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository()
	collaboratorRepo := repository.NewCollaboratorRepository()
	participantRepo := repository.NewParticipantRepository()
	achievementRepo := repository.NewUserAggregateRepository()
	oauth2Repo := repository.NewOAuth2Repository()
	userRepo := repository.NewUserRepository()
	projectRepo := repository.NewProjectRepository()

	autoTextQuest := &entity.Quest{
		Base:           entity.Base{ID: "auto text quest"},
		ProjectID:      testutil.Project1.ID,
		Type:           entity.QuestText,
		Status:         entity.QuestActive,
		CategoryIDs:    []string{},
		Recurrence:     entity.Daily,
		ValidationData: entity.Map{"auto_validate": true, "answer": "Foo"},
		ConditionOp:    entity.Or,
	}

	err := questRepo.Create(ctx, autoTextQuest)
	require.NoError(t, err)

	d := NewClaimedQuestDomain(
		claimedQuestRepo,
		questRepo,
		collaboratorRepo,
		participantRepo,
		oauth2Repo,
		achievementRepo,
		userRepo,
		projectRepo,
		&testutil.MockTwitterEndpoint{},
		&testutil.MockDiscordEndpoint{},
	)

	// User1 cannot claim quest with a wrong answer.
	authorizedCtx := testutil.NewMockContextWithUserID(ctx, testutil.User1.ID)
	resp, err := d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID: autoTextQuest.ID,
		Input:   "wrong answer",
	})
	require.NoError(t, err)
	require.Equal(t, "auto_rejected", resp.Status)

	// User1 claims quest again but with a correct answer.
	authorizedCtx = testutil.NewMockContextWithUserID(ctx, testutil.User1.ID)
	resp, err = d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID: autoTextQuest.ID,
		Input:   "Foo",
	})
	require.NoError(t, err)
	require.Equal(t, "auto_accepted", resp.Status)

	// User1 cannot claims quest again because the daily recurrence.
	authorizedCtx = testutil.NewMockContextWithUserID(ctx, testutil.User1.ID)
	_, err = d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID: autoTextQuest.ID,
		Input:   "Foo",
	})
	require.Error(t, err)
	require.Equal(t, "This quest cannot be claimed now", err.Error())
}

func Test_claimedQuestDomain_Claim_GivePoint(t *testing.T) {
	ctx := testutil.NewMockContext()
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository()
	collaboratorRepo := repository.NewCollaboratorRepository()
	participantRepo := repository.NewParticipantRepository()
	achievementRepo := repository.NewUserAggregateRepository()
	oauth2Repo := repository.NewOAuth2Repository()
	userRepo := repository.NewUserRepository()
	projectRepo := repository.NewProjectRepository()

	autoTextQuest := &entity.Quest{
		Base:           entity.Base{ID: "auto text quest"},
		ProjectID:      testutil.Project1.ID,
		Type:           entity.QuestText,
		Status:         entity.QuestActive,
		CategoryIDs:    []string{},
		Recurrence:     entity.Daily,
		ValidationData: entity.Map{"auto_validate": true, "answer": "Foo"},
		ConditionOp:    entity.Or,
		Rewards:        []entity.Reward{{Type: entity.PointReward, Data: entity.Map{"points": 100}}},
	}

	err := questRepo.Create(ctx, autoTextQuest)
	require.NoError(t, err)

	d := NewClaimedQuestDomain(
		claimedQuestRepo,
		questRepo,
		collaboratorRepo,
		participantRepo,
		oauth2Repo,
		achievementRepo,
		userRepo,
		projectRepo,
		&testutil.MockTwitterEndpoint{},
		&testutil.MockDiscordEndpoint{},
	)

	// User claims the quest.
	authorizedCtx := testutil.NewMockContextWithUserID(ctx, testutil.User1.ID)
	resp, err := d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID: autoTextQuest.ID,
		Input:   "Foo",
	})
	require.NoError(t, err)
	require.Equal(t, "auto_accepted", resp.Status)

	// Check points from participant repo.
	participant, err := participantRepo.Get(ctx, testutil.User1.ID, autoTextQuest.ProjectID)
	require.NoError(t, err)
	require.Equal(t, uint64(100), participant.Points)
}

func Test_claimedQuestDomain_Claim_ManualText(t *testing.T) {
	ctx := testutil.NewMockContext()
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository()
	collaboratorRepo := repository.NewCollaboratorRepository()
	participantRepo := repository.NewParticipantRepository()
	achievementRepo := repository.NewUserAggregateRepository()
	oauth2Repo := repository.NewOAuth2Repository()
	userRepo := repository.NewUserRepository()
	projectRepo := repository.NewProjectRepository()

	autoTextQuest := &entity.Quest{
		Base:           entity.Base{ID: "manual text quest"},
		ProjectID:      testutil.Project1.ID,
		Type:           entity.QuestText,
		Status:         entity.QuestActive,
		CategoryIDs:    []string{},
		Recurrence:     entity.Daily,
		ValidationData: entity.Map{"auto_validate": false},
		ConditionOp:    entity.Or,
	}

	err := questRepo.Create(ctx, autoTextQuest)
	require.NoError(t, err)

	d := NewClaimedQuestDomain(
		claimedQuestRepo,
		questRepo,
		collaboratorRepo,
		participantRepo,
		oauth2Repo,
		achievementRepo,
		userRepo,
		projectRepo,
		&testutil.MockTwitterEndpoint{},
		&testutil.MockDiscordEndpoint{},
	)

	// Need to wait for a manual review if user claims a manual text quest.
	authorizedCtx := testutil.NewMockContextWithUserID(ctx, testutil.User1.ID)
	got, err := d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID: autoTextQuest.ID,
		Input:   "any anwser",
	})
	require.NoError(t, err)
	require.Equal(t, "pending", got.Status)

	// Cannot claim the quest again while the quest is pending.
	authorizedCtx = testutil.NewMockContextWithUserID(ctx, testutil.User1.ID)
	_, err = d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID: autoTextQuest.ID,
		Input:   "any anwser",
	})
	require.Error(t, err)
	require.Equal(t, "This quest cannot be claimed now", err.Error())
}

func Test_claimedQuestDomain_Claim_CreateUserAggregate(t *testing.T) {
	ctx := testutil.NewMockContext()
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository()
	collaboratorRepo := repository.NewCollaboratorRepository()
	participantRepo := repository.NewParticipantRepository()
	achievementRepo := repository.NewUserAggregateRepository()
	oauth2Repo := repository.NewOAuth2Repository()
	userRepo := repository.NewUserRepository()
	projectRepo := repository.NewProjectRepository()

	d := NewClaimedQuestDomain(
		claimedQuestRepo,
		questRepo,
		collaboratorRepo,
		participantRepo,
		oauth2Repo,
		achievementRepo,
		userRepo,
		projectRepo,
		&testutil.MockTwitterEndpoint{},
		&testutil.MockDiscordEndpoint{},
	)

	// User claims the quest.
	authorizedCtx := testutil.NewMockContextWithUserID(ctx, testutil.User1.ID)
	resp, err := d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID: testutil.Quest3.ID,
		Input:   "any",
	})

	require.NoError(t, err)
	require.Equal(t, "auto_accepted", resp.Status)

	expected := []*entity.UserAggregate{
		{
			ProjectID:  testutil.Quest1.ProjectID,
			UserID:     testutil.User1.ID,
			Range:      entity.UserAggregateRangeMonth,
			TotalTask:  1,
			TotalPoint: 100,
		},
		{
			ProjectID:  testutil.Quest1.ProjectID,
			UserID:     testutil.User1.ID,
			Range:      entity.UserAggregateRangeWeek,
			TotalTask:  1,
			TotalPoint: 100,
		},
		{
			ProjectID:  testutil.Quest1.ProjectID,
			UserID:     testutil.User1.ID,
			Range:      entity.UserAggregateRangeTotal,
			TotalTask:  1,
			TotalPoint: 100,
		},
	}

	var actual []*entity.UserAggregate
	tx := ctx.DB().Model(&entity.UserAggregate{}).Where("project_id = ?", testutil.Quest1.ProjectID).Find(&actual)
	require.NoError(t, tx.Error)

	require.Equal(t, 3, len(actual))

	sort.SliceStable(actual, func(i, j int) bool {
		return actual[i].Range < actual[j].Range
	})

	sort.SliceStable(expected, func(i, j int) bool {
		return expected[i].Range < expected[j].Range
	})

	for i := 0; i < len(actual); i++ {
		require.True(t, reflectutil.PartialEqual(expected[i], actual[i]))
	}
}

func Test_claimedQuestDomain_Claim(t *testing.T) {
	type args struct {
		ctx xcontext.Context
		req *model.ClaimQuestRequest
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "cannot claim draft quest",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User1.ID),
				req: &model.ClaimQuestRequest{
					QuestID: testutil.Quest1.ID,
					Input:   "Bar",
				},
			},
			wantErr: errors.New("Only allow to claim active quests"),
		},
		{
			name: "cannot claim quest2 if user has not claimed quest1 yet",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User2.ID),
				req: &model.ClaimQuestRequest{
					QuestID: testutil.Quest2.ID,
				},
			},
			wantErr: errors.New("This quest cannot be claimed now"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			d := NewClaimedQuestDomain(
				repository.NewClaimedQuestRepository(),
				repository.NewQuestRepository(),
				repository.NewCollaboratorRepository(),
				repository.NewParticipantRepository(),
				repository.NewOAuth2Repository(),
				repository.NewUserAggregateRepository(),
				repository.NewUserRepository(),
				repository.NewProjectRepository(),
				&testutil.MockTwitterEndpoint{},
				&testutil.MockDiscordEndpoint{},
			)

			got, err := d.Claim(tt.args.ctx, tt.args.req)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("newVisitLinkValidator() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_claimedQuestDomain_Get(t *testing.T) {
	type args struct {
		ctx xcontext.Context
		req *model.GetClaimedQuestRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.GetClaimedQuestResponse
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Collaborator1.UserID),
				req: &model.GetClaimedQuestRequest{
					ID: testutil.ClaimedQuest1.ID,
				},
			},
			want: &model.GetClaimedQuestResponse{
				QuestID:    testutil.ClaimedQuest1.QuestID,
				UserID:     testutil.ClaimedQuest1.UserID,
				Input:      testutil.ClaimedQuest1.Input,
				Status:     string(testutil.ClaimedQuest1.Status),
				ReviewerID: testutil.ClaimedQuest1.ReviewerID,
				ReviewerAt: testutil.ClaimedQuest1.ReviewerAt.Format(time.RFC3339Nano),
				Comment:    testutil.ClaimedQuest1.Comment,
			},
			wantErr: nil,
		},
		{
			name: "invalid id",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Collaborator1.UserID),
				req: &model.GetClaimedQuestRequest{
					ID: "invalid id",
				},
			},
			want:    nil,
			wantErr: errors.New("Not found claimed quest"),
		},
		{
			name: "permission denied",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User2.ID),
				req: &model.GetClaimedQuestRequest{
					ID: testutil.ClaimedQuest1.ID,
				},
			},
			want:    nil,
			wantErr: errors.New("Permission denied"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			d := &claimedQuestDomain{
				claimedQuestRepo: repository.NewClaimedQuestRepository(),
				questRepo:        repository.NewQuestRepository(),
				roleVerifier:     common.NewProjectRoleVerifier(repository.NewCollaboratorRepository(), repository.NewUserRepository()),
			}

			got, err := d.Get(tt.args.ctx, tt.args.req)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.True(t, reflectutil.PartialEqual(tt.want, got), "%v != %v", tt.want, got)
			}
		})
	}
}

func Test_claimedQuestDomain_GetList(t *testing.T) {
	type args struct {
		ctx xcontext.Context
		req *model.GetListClaimedQuestRequest
	}

	tests := []struct {
		name    string
		args    args
		want    *model.GetListClaimedQuestResponse
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    0,
					Limit:     2,
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						ID:      testutil.ClaimedQuest1.ID,
						QuestID: testutil.ClaimedQuest1.QuestID,
						Quest: model.Quest{
							ID:             testutil.Quest1.ID,
							ProjectID:      testutil.Quest1.ProjectID,
							Type:           string(testutil.Quest1.Type),
							Status:         string(testutil.Quest1.Status),
							Title:          testutil.Quest1.Title,
							Description:    string(testutil.Quest1.Description),
							Categories:     testutil.Quest1.CategoryIDs,
							Recurrence:     string(testutil.Quest1.Recurrence),
							ValidationData: testutil.Quest1.ValidationData,
							Rewards:        rewardEntityToModel(testutil.Quest1.Rewards),
							ConditionOp:    string(testutil.Quest1.ConditionOp),
							Conditions:     conditionEntityToModel(testutil.Quest1.Conditions),
						},
						UserID: testutil.ClaimedQuest1.UserID,
						User: model.User{
							ID: testutil.User1.ID,
						},
						Status:     string(testutil.ClaimedQuest1.Status),
						ReviewerID: testutil.ClaimedQuest1.ReviewerID,
						ReviewerAt: testutil.ClaimedQuest1.ReviewerAt.Format(time.RFC3339Nano),
					},
					{
						ID:         testutil.ClaimedQuest2.ID,
						QuestID:    testutil.ClaimedQuest2.QuestID,
						UserID:     testutil.ClaimedQuest2.UserID,
						Status:     string(testutil.ClaimedQuest2.Status),
						ReviewerID: testutil.ClaimedQuest2.ReviewerID,
						ReviewerAt: testutil.ClaimedQuest2.ReviewerAt.Format(time.RFC3339Nano),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case with custom offset",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    2,
					Limit:     1,
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						ID:         testutil.ClaimedQuest3.ID,
						QuestID:    testutil.ClaimedQuest3.QuestID,
						UserID:     testutil.ClaimedQuest3.UserID,
						Status:     string(testutil.ClaimedQuest3.Status),
						ReviewerID: testutil.ClaimedQuest3.ReviewerID,
						ReviewerAt: testutil.ClaimedQuest3.ReviewerAt.Format(time.RFC3339Nano),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "nagative limit",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    2,
					Limit:     -1,
				},
			},
			want:    nil,
			wantErr: errors.New("Limit must be positive"),
		},
		{
			name: "exceed the maximum limit",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    2,
					Limit:     51,
				},
			},
			want:    nil,
			wantErr: errors.New("Exceed the maximum of limit"),
		},
		{
			name: "permission denied",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User2.ID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    2,
					Limit:     51,
				},
			},
			want:    nil,
			wantErr: errors.New("Permission denied"),
		},
		{
			name: "filter by accepted",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID:    testutil.Project1.ID,
					FilterStatus: string(entity.Accepted),
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						ID:         testutil.ClaimedQuest1.ID,
						QuestID:    testutil.ClaimedQuest1.QuestID,
						UserID:     testutil.ClaimedQuest1.UserID,
						Status:     string(testutil.ClaimedQuest1.Status),
						ReviewerID: testutil.ClaimedQuest1.ReviewerID,
						ReviewerAt: testutil.ClaimedQuest1.ReviewerAt.Format(time.RFC3339Nano),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "filter by rejected",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID:    testutil.Project1.ID,
					FilterStatus: string(entity.Rejected),
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						ID:         testutil.ClaimedQuest2.ID,
						QuestID:    testutil.ClaimedQuest2.QuestID,
						UserID:     testutil.ClaimedQuest2.UserID,
						Status:     string(testutil.ClaimedQuest2.Status),
						ReviewerID: testutil.ClaimedQuest2.ReviewerID,
						ReviewerAt: testutil.ClaimedQuest2.ReviewerAt.Format(time.RFC3339Nano),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "filter by quest and pending",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID:     testutil.Project1.ID,
					FilterStatus:  string(entity.Pending),
					FilterQuestID: testutil.ClaimedQuest3.QuestID,
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						ID:         testutil.ClaimedQuest3.ID,
						QuestID:    testutil.ClaimedQuest3.QuestID,
						UserID:     testutil.ClaimedQuest3.UserID,
						Status:     string(testutil.ClaimedQuest3.Status),
						ReviewerID: testutil.ClaimedQuest3.ReviewerID,
						ReviewerAt: testutil.ClaimedQuest3.ReviewerAt.Format(time.RFC3339Nano),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "filter by user and pending",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID:    testutil.Project1.ID,
					FilterStatus: string(entity.Pending),
					FilterUserID: testutil.ClaimedQuest3.UserID,
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						ID:         testutil.ClaimedQuest3.ID,
						QuestID:    testutil.ClaimedQuest3.QuestID,
						UserID:     testutil.ClaimedQuest3.UserID,
						Status:     string(testutil.ClaimedQuest3.Status),
						ReviewerID: testutil.ClaimedQuest3.ReviewerID,
						ReviewerAt: testutil.ClaimedQuest3.ReviewerAt.Format(time.RFC3339Nano),
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			d := &claimedQuestDomain{
				claimedQuestRepo: repository.NewClaimedQuestRepository(),
				questRepo:        repository.NewQuestRepository(),
				userRepo:         repository.NewUserRepository(),
				roleVerifier:     common.NewProjectRoleVerifier(repository.NewCollaboratorRepository(), repository.NewUserRepository()),
			}

			got, err := d.GetList(tt.args.ctx, tt.args.req)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.True(t, reflectutil.PartialEqual(tt.want, got), "%v != %v", tt.want, got)
			}
		})
	}
}

func Test_claimedQuestDomain_ReviewClaimedQuest(t *testing.T) {
	type args struct {
		ctx xcontext.Context
		req *model.ReviewClaimedQuestRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.ReviewClaimedQuestResponse
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User3.ID),
				req: &model.ReviewClaimedQuestRequest{
					ID:     testutil.ClaimedQuest3.ID,
					Action: string(entity.Accepted),
				},
			},
			want: &model.ReviewClaimedQuestResponse{},
		},
		{
			name: "err claimed quest must be pending",
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.User1.ID),
				req: &model.ReviewClaimedQuestRequest{
					ID:     testutil.ClaimedQuest1.ID,
					Action: string(entity.Accepted),
				},
			},
			wantErr: errorx.New(errorx.BadRequest, "Claimed quest must be pending"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			d := NewClaimedQuestDomain(
				repository.NewClaimedQuestRepository(),
				repository.NewQuestRepository(),
				repository.NewCollaboratorRepository(),
				repository.NewParticipantRepository(),
				repository.NewOAuth2Repository(),
				repository.NewUserAggregateRepository(),
				repository.NewUserRepository(),
				repository.NewProjectRepository(),
				&testutil.MockTwitterEndpoint{},
				&testutil.MockDiscordEndpoint{},
			)

			got, err := d.ReviewClaimedQuest(tt.args.ctx, tt.args.req)
			if err != nil && err != tt.wantErr {
				t.Errorf("claimedQuestDomain.ReviewClaimedQuest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("claimedQuestDomain.ReviewClaimedQuest() = %v, want %v", got, tt.want)
			}
		})
	}
}
