package domain

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/badge"
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
	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	collaboratorRepo := repository.NewCollaboratorRepository()
	followerRepo := repository.NewFollowerRepository()
	oauth2Repo := repository.NewOAuth2Repository()
	userRepo := repository.NewUserRepository()
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{})
	payRewardRepo := repository.NewPayRewardRepository()
	categoryRepo := repository.NewCategoryRepository()

	autoTextQuest := &entity.Quest{
		Base:           entity.Base{ID: "auto text quest"},
		CommunityID:    sql.NullString{Valid: true, String: testutil.Community1.ID},
		Type:           entity.QuestText,
		Status:         entity.QuestActive,
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
		followerRepo,
		oauth2Repo,
		userRepo,
		communityRepo,
		payRewardRepo,
		categoryRepo,
		&testutil.MockTwitterEndpoint{},
		&testutil.MockDiscordEndpoint{},
		nil,
		badge.NewManager(
			repository.NewBadgeRepository(),
			badge.NewRainBowBadgeScanner(followerRepo, []uint64{1}),
			badge.NewQuestWarriorBadgeScanner(followerRepo, []uint64{1}),
		),
		&testutil.MockLeaderboard{},
	)

	// User1 cannot claim quest with a wrong answer.
	authorizedCtx := xcontext.WithRequestUserID(ctx, testutil.User1.ID)
	resp, err := d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID:        autoTextQuest.ID,
		SubmissionData: "wrong answer",
	})
	require.NoError(t, err)
	require.Equal(t, "auto_rejected", resp.Status)
	require.Equal(t, "Wrong answer", resp.Message)

	// User1 claims quest again but with a correct answer.
	authorizedCtx = xcontext.WithRequestUserID(ctx, testutil.User1.ID)
	resp, err = d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID:        autoTextQuest.ID,
		SubmissionData: "Foo",
	})
	require.NoError(t, err)
	require.Equal(t, "auto_accepted", resp.Status)

	// User1 cannot claims quest again because the daily recurrence.
	authorizedCtx = xcontext.WithRequestUserID(ctx, testutil.User1.ID)
	_, err = d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID:        autoTextQuest.ID,
		SubmissionData: "Foo",
	})
	require.Error(t, err)
	require.Equal(t, "Please wait until the next day to claim this quest", err.Error())
}

func Test_claimedQuestDomain_Claim_GivePoint(t *testing.T) {
	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	collaboratorRepo := repository.NewCollaboratorRepository()
	followerRepo := repository.NewFollowerRepository()
	oauth2Repo := repository.NewOAuth2Repository()
	userRepo := repository.NewUserRepository()
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{})
	badgeRepo := repository.NewBadgeRepository()
	payRewardRepo := repository.NewPayRewardRepository()
	categoryRepo := repository.NewCategoryRepository()

	autoTextQuest := &entity.Quest{
		Base:           entity.Base{ID: "auto text quest"},
		CommunityID:    sql.NullString{Valid: true, String: testutil.Community2.ID},
		Type:           entity.QuestText,
		Status:         entity.QuestActive,
		Recurrence:     entity.Daily,
		ValidationData: entity.Map{"auto_validate": true, "answer": "Foo"},
		ConditionOp:    entity.Or,
		Points:         100,
	}

	err := questRepo.Create(ctx, autoTextQuest)
	require.NoError(t, err)

	d := NewClaimedQuestDomain(
		claimedQuestRepo,
		questRepo,
		collaboratorRepo,
		followerRepo,
		oauth2Repo,
		userRepo,
		communityRepo,
		payRewardRepo,
		categoryRepo,
		&testutil.MockTwitterEndpoint{},
		&testutil.MockDiscordEndpoint{},
		nil,
		badge.NewManager(
			badgeRepo,
			badge.NewRainBowBadgeScanner(followerRepo, []uint64{1}),
			badge.NewQuestWarriorBadgeScanner(followerRepo, []uint64{1}),
		),
		&testutil.MockLeaderboard{},
	)

	// User claims the quest.
	authorizedCtx := xcontext.WithRequestUserID(ctx, testutil.User1.ID)
	resp, err := d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID:        autoTextQuest.ID,
		SubmissionData: "Foo",
	})
	require.NoError(t, err)
	require.Equal(t, "auto_accepted", resp.Status)

	// Check points from follower repo.
	follower, err := followerRepo.Get(ctx, testutil.User1.ID, autoTextQuest.CommunityID.String)
	require.NoError(t, err)
	require.Equal(t, uint64(100), follower.Points)
	require.Equal(t, uint64(1), follower.Streaks)

	// Check rainbow (streak) badge.
	myBadge, err := badgeRepo.Get(
		ctx,
		testutil.User1.ID, autoTextQuest.CommunityID.String,
		badge.RainBowBadgeName,
	)
	require.NoError(t, err)
	require.Equal(t, 1, myBadge.Level)

	// Check quest warrior badge.
	myBadge, err = badgeRepo.Get(
		ctx,
		testutil.User1.ID,
		autoTextQuest.CommunityID.String,
		badge.QuestWarriorBadgeName,
	)
	require.NoError(t, err)
	require.Equal(t, 1, myBadge.Level)
}

func Test_claimedQuestDomain_Claim_ManualText(t *testing.T) {
	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	collaboratorRepo := repository.NewCollaboratorRepository()
	followerRepo := repository.NewFollowerRepository()
	oauth2Repo := repository.NewOAuth2Repository()
	userRepo := repository.NewUserRepository()
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{})
	transactionRepo := repository.NewPayRewardRepository()
	categoryRepo := repository.NewCategoryRepository()

	autoTextQuest := &entity.Quest{
		Base:           entity.Base{ID: "manual text quest"},
		CommunityID:    sql.NullString{Valid: true, String: testutil.Community1.ID},
		Type:           entity.QuestText,
		Status:         entity.QuestActive,
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
		followerRepo,
		oauth2Repo,
		userRepo,
		communityRepo,
		transactionRepo,
		categoryRepo,
		&testutil.MockTwitterEndpoint{},
		&testutil.MockDiscordEndpoint{},
		nil,
		badge.NewManager(
			repository.NewBadgeRepository(),
			badge.NewRainBowBadgeScanner(followerRepo, []uint64{1}),
			badge.NewQuestWarriorBadgeScanner(followerRepo, []uint64{1}),
		),
		&testutil.MockLeaderboard{},
	)

	// Need to wait for a manual review if user claims a manual text quest.
	authorizedCtx := xcontext.WithRequestUserID(ctx, testutil.User1.ID)
	got, err := d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID:        autoTextQuest.ID,
		SubmissionData: "any anwser",
	})
	require.NoError(t, err)
	require.Equal(t, "pending", got.Status)

	// Cannot claim the quest again while the quest is pending.
	authorizedCtx = xcontext.WithRequestUserID(ctx, testutil.User1.ID)
	_, err = d.Claim(authorizedCtx, &model.ClaimQuestRequest{
		QuestID:        autoTextQuest.ID,
		SubmissionData: "any anwser",
	})
	require.Error(t, err)
	require.Equal(t, "Please wait until the next day to claim this quest", err.Error())
}

func Test_claimedQuestDomain_Claim(t *testing.T) {
	type args struct {
		ctx context.Context
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
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.ClaimQuestRequest{
					QuestID:        testutil.Quest1.ID,
					SubmissionData: "Bar",
				},
			},
			wantErr: errors.New("Only allow to claim active quests"),
		},
		{
			name: "cannot claim quest2 if user has not claimed quest1 yet",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User2.ID),
				req: &model.ClaimQuestRequest{
					QuestID: testutil.Quest2.ID,
				},
			},
			wantErr: errors.New("Please complete quest Quest 1 before claiming this quest"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			d := NewClaimedQuestDomain(
				repository.NewClaimedQuestRepository(),
				repository.NewQuestRepository(&testutil.MockSearchCaller{}),
				repository.NewCollaboratorRepository(),
				repository.NewFollowerRepository(),
				repository.NewOAuth2Repository(),
				repository.NewUserRepository(),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
				repository.NewPayRewardRepository(),
				repository.NewCategoryRepository(),
				&testutil.MockTwitterEndpoint{},
				&testutil.MockDiscordEndpoint{},
				nil,
				badge.NewManager(repository.NewBadgeRepository()),
				&testutil.MockLeaderboard{},
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
		ctx context.Context
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
				ctx: testutil.MockContextWithUserID(testutil.Collaborator1.UserID),
				req: &model.GetClaimedQuestRequest{
					ID: testutil.ClaimedQuest1.ID,
				},
			},
			want: &model.GetClaimedQuestResponse{
				Quest: model.Quest{
					ID:          testutil.Quest1.ID,
					Community:   model.Community{Handle: testutil.Community1.Handle},
					Type:        string(testutil.Quest1.Type),
					Status:      string(testutil.Quest1.Status),
					Title:       testutil.Quest1.Title,
					Description: string(testutil.Quest1.Description),
					Category: model.Category{
						ID:   testutil.Category1.ID,
						Name: testutil.Category1.Name,
					},
					Recurrence:     string(testutil.Quest1.Recurrence),
					ValidationData: testutil.Quest1.ValidationData,
					Rewards:        convertRewards(testutil.Quest1.Rewards),
					ConditionOp:    string(testutil.Quest1.ConditionOp),
					Conditions:     convertConditions(testutil.Quest1.Conditions),
				},
				User: model.User{
					ID: testutil.User1.ID,
				},
				SubmissionData: testutil.ClaimedQuest1.SubmissionData,
				Status:         string(testutil.ClaimedQuest1.Status),
				ReviewerID:     testutil.ClaimedQuest1.ReviewerID,
				Comment:        testutil.ClaimedQuest1.Comment,
			},
			wantErr: nil,
		},
		{
			name: "invalid id",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Collaborator1.UserID),
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
				ctx: testutil.MockContextWithUserID(testutil.User2.ID),
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
				questRepo:        repository.NewQuestRepository(&testutil.MockSearchCaller{}),
				userRepo:         repository.NewUserRepository(),
				categoryRepo:     repository.NewCategoryRepository(),
				communityRepo:    repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
				roleVerifier:     common.NewCommunityRoleVerifier(repository.NewCollaboratorRepository(), repository.NewUserRepository()),
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
		ctx context.Context
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
				ctx: testutil.MockContextWithUserID(testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Offset:          0,
					Limit:           2,
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						ID: testutil.ClaimedQuest1.ID,
						Quest: model.Quest{
							ID:          testutil.Quest1.ID,
							Community:   model.Community{Handle: testutil.Community1.Handle},
							Type:        string(testutil.Quest1.Type),
							Status:      string(testutil.Quest1.Status),
							Title:       testutil.Quest1.Title,
							Description: string(testutil.Quest1.Description),
							Category: model.Category{
								ID:   testutil.Category1.ID,
								Name: testutil.Category1.Name,
							},
							Recurrence:     string(testutil.Quest1.Recurrence),
							ValidationData: testutil.Quest1.ValidationData,
							Rewards:        convertRewards(testutil.Quest1.Rewards),
							ConditionOp:    string(testutil.Quest1.ConditionOp),
							Conditions:     convertConditions(testutil.Quest1.Conditions),
						},
						User: model.User{
							ID: testutil.User1.ID,
						},
						Status:     string(testutil.ClaimedQuest1.Status),
						ReviewerID: testutil.ClaimedQuest1.ReviewerID,
					},
					{
						ID:         testutil.ClaimedQuest2.ID,
						Quest:      model.Quest{ID: testutil.ClaimedQuest2.QuestID},
						User:       model.User{ID: testutil.ClaimedQuest2.UserID},
						Status:     string(testutil.ClaimedQuest2.Status),
						ReviewerID: testutil.ClaimedQuest2.ReviewerID,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case with custom offset",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Offset:          2,
					Limit:           1,
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						ID:         testutil.ClaimedQuest3.ID,
						Quest:      model.Quest{ID: testutil.ClaimedQuest3.QuestID},
						User:       model.User{ID: testutil.ClaimedQuest3.UserID},
						Status:     string(testutil.ClaimedQuest3.Status),
						ReviewerID: testutil.ClaimedQuest3.ReviewerID,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "nagative limit",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Offset:          2,
					Limit:           -1,
				},
			},
			want:    nil,
			wantErr: errors.New("Limit must be positive"),
		},
		{
			name: "exceed the maximum limit",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Offset:          2,
					Limit:           51,
				},
			},
			want:    nil,
			wantErr: errors.New("Exceed the maximum of limit (50)"),
		},
		{
			name: "permission denied",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User2.ID),
				req: &model.GetListClaimedQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Offset:          2,
					Limit:           51,
				},
			},
			want:    nil,
			wantErr: errors.New("Permission denied"),
		},
		{
			name: "filter by accepted",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Status:          string(entity.Accepted),
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						ID:         testutil.ClaimedQuest1.ID,
						Quest:      model.Quest{ID: testutil.ClaimedQuest1.QuestID},
						User:       model.User{ID: testutil.ClaimedQuest1.UserID},
						Status:     string(testutil.ClaimedQuest1.Status),
						ReviewerID: testutil.ClaimedQuest1.ReviewerID,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "filter by rejected",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Status:          string(entity.Rejected),
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						ID:         testutil.ClaimedQuest2.ID,
						Quest:      model.Quest{ID: testutil.ClaimedQuest2.QuestID},
						User:       model.User{ID: testutil.ClaimedQuest2.UserID},
						Status:     string(testutil.ClaimedQuest2.Status),
						ReviewerID: testutil.ClaimedQuest2.ReviewerID,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "filter by quest and pending",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Status:          string(entity.Pending),
					QuestID:         testutil.ClaimedQuest3.QuestID,
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						ID:         testutil.ClaimedQuest3.ID,
						Quest:      model.Quest{ID: testutil.ClaimedQuest3.QuestID},
						User:       model.User{ID: testutil.ClaimedQuest3.UserID},
						Status:     string(testutil.ClaimedQuest3.Status),
						ReviewerID: testutil.ClaimedQuest3.ReviewerID,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "filter by user and pending",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.Collaborator1.UserID),
				req: &model.GetListClaimedQuestRequest{
					CommunityHandle: testutil.Community1.Handle,
					Status:          string(entity.Pending),
					UserID:          testutil.ClaimedQuest3.UserID,
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						ID:         testutil.ClaimedQuest3.ID,
						Quest:      model.Quest{ID: testutil.ClaimedQuest3.QuestID},
						User:       model.User{ID: testutil.ClaimedQuest3.UserID},
						Status:     string(testutil.ClaimedQuest3.Status),
						ReviewerID: testutil.ClaimedQuest3.ReviewerID,
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
				questRepo:        repository.NewQuestRepository(&testutil.MockSearchCaller{}),
				userRepo:         repository.NewUserRepository(),
				categoryRepo:     repository.NewCategoryRepository(),
				communityRepo:    repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
				roleVerifier:     common.NewCommunityRoleVerifier(repository.NewCollaboratorRepository(), repository.NewUserRepository()),
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

func Test_claimedQuestDomain_Review(t *testing.T) {
	type args struct {
		ctx context.Context
		req *model.ReviewRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.ReviewResponse
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User3.ID),
				req: &model.ReviewRequest{
					IDs:    []string{testutil.ClaimedQuest3.ID},
					Action: string(entity.Accepted),
				},
			},
			want: &model.ReviewResponse{},
		},
		{
			name: "err claimed quest must be pending",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.ReviewRequest{
					IDs:    []string{testutil.ClaimedQuest1.ID},
					Action: string(entity.Accepted),
				},
			},
			wantErr: errorx.New(errorx.BadRequest, "Claimed quest must be pending"),
		},
		{
			name: "permission denied",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User2.ID),
				req: &model.ReviewRequest{
					IDs:    []string{testutil.ClaimedQuest3.ID},
					Action: string(entity.Accepted),
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			d := NewClaimedQuestDomain(
				repository.NewClaimedQuestRepository(),
				repository.NewQuestRepository(&testutil.MockSearchCaller{}),
				repository.NewCollaboratorRepository(),
				repository.NewFollowerRepository(),
				repository.NewOAuth2Repository(),
				repository.NewUserRepository(),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
				repository.NewPayRewardRepository(),
				repository.NewCategoryRepository(),
				&testutil.MockTwitterEndpoint{},
				&testutil.MockDiscordEndpoint{},
				nil,
				badge.NewManager(
					repository.NewBadgeRepository(),
					badge.NewQuestWarriorBadgeScanner(repository.NewFollowerRepository(), []uint64{1}),
				),
				&testutil.MockLeaderboard{},
			)

			got, err := d.Review(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

			if tt.want != nil {
				require.True(t, reflect.DeepEqual(got, tt.want), "%v != %v", got, tt.want)
			}
		})
	}
}

func Test_claimedQuestDomain_ReviewAll(t *testing.T) {
	type args struct {
		ctx context.Context
		req *model.ReviewAllRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.ReviewAllResponse
		wantErr error
	}{
		{
			name: "happy case filter by quest",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User3.ID),
				req: &model.ReviewAllRequest{
					Action:          string(entity.Accepted),
					CommunityHandle: testutil.Community1.Handle,
					QuestIDs:        []string{testutil.Quest1.ID},
				},
			},
			want: &model.ReviewAllResponse{Quantity: 2},
		},
		{
			name: "happy case filter by user",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User3.ID),
				req: &model.ReviewAllRequest{
					Action:          string(entity.Accepted),
					CommunityHandle: testutil.Community1.Handle,
					UserIDs:         []string{testutil.User2.ID},
				},
			},
			want: &model.ReviewAllResponse{Quantity: 1},
		},
		{
			name: "happy case with excludes",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.ReviewAllRequest{
					Action:          string(entity.Accepted),
					CommunityHandle: testutil.Community1.Handle,
					QuestIDs:        []string{testutil.Quest1.ID},
					Excludes:        []string{"claimed_quest_test_1"},
				},
			},
			want: &model.ReviewAllResponse{Quantity: 1},
		},
		{
			name: "invalid status",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.ReviewAllRequest{
					Action:          "invalid",
					CommunityHandle: testutil.Community1.Handle,
					QuestIDs:        []string{testutil.Quest1.ID},
					Excludes:        []string{"claimed_quest_test_1"},
				},
			},
			wantErr: errorx.New(errorx.BadRequest, "Invalid action"),
		},
		{
			name: "permission denied",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User2.ID),
				req: &model.ReviewAllRequest{
					Action:          string(entity.Accepted),
					CommunityHandle: testutil.Community1.Handle,
					QuestIDs:        []string{testutil.Quest1.ID},
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
		{
			name: "unapprove a claimed quest",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.ReviewAllRequest{
					Action:          string(entity.Pending),
					CommunityHandle: testutil.Community1.Handle,
					QuestIDs:        []string{testutil.Quest3.ID},
					Statuses:        []string{string(entity.Accepted)},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			claimedQuestRepo := repository.NewClaimedQuestRepository()
			err := claimedQuestRepo.Create(tt.args.ctx, &entity.ClaimedQuest{
				Base:    entity.Base{ID: "claimed_quest_test_1"},
				QuestID: testutil.Quest1.ID,
				UserID:  testutil.User2.ID,
				Status:  entity.Pending,
			})
			require.NoError(t, err)

			err = claimedQuestRepo.Create(tt.args.ctx, &entity.ClaimedQuest{
				Base:    entity.Base{ID: "claimed_quest_test_2"},
				QuestID: testutil.Quest1.ID,
				UserID:  testutil.User3.ID,
				Status:  entity.Pending,
			})
			require.NoError(t, err)

			err = claimedQuestRepo.Create(tt.args.ctx, &entity.ClaimedQuest{
				Base:    entity.Base{ID: "claimed_quest_test_3"},
				QuestID: testutil.Quest2.ID,
				UserID:  testutil.User1.ID,
				Status:  entity.Pending,
			})
			require.NoError(t, err)

			err = claimedQuestRepo.Create(tt.args.ctx, &entity.ClaimedQuest{
				Base:    entity.Base{ID: "claimed_quest_test_4"},
				QuestID: testutil.Quest3.ID,
				UserID:  testutil.User1.ID,
				Status:  entity.Accepted,
			})
			require.NoError(t, err)

			d := NewClaimedQuestDomain(
				repository.NewClaimedQuestRepository(),
				repository.NewQuestRepository(&testutil.MockSearchCaller{}),
				repository.NewCollaboratorRepository(),
				repository.NewFollowerRepository(),
				repository.NewOAuth2Repository(),
				repository.NewUserRepository(),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}),
				repository.NewPayRewardRepository(),
				repository.NewCategoryRepository(),
				&testutil.MockTwitterEndpoint{},
				&testutil.MockDiscordEndpoint{},
				nil,
				badge.NewManager(
					repository.NewBadgeRepository(),
					badge.NewQuestWarriorBadgeScanner(repository.NewFollowerRepository(), []uint64{1}),
				),
				&testutil.MockLeaderboard{},
			)

			got, err := d.ReviewAll(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, err, tt.wantErr)
			}

			if tt.want != nil {
				require.True(t, reflect.DeepEqual(got, tt.want), "%v != %v", got, tt.want)
			}
		})
	}
}

func Test_fullScenario_ClaimInvitedCommunity(t *testing.T) {
	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	collaboratorRepo := repository.NewCollaboratorRepository()
	followerRepo := repository.NewFollowerRepository()
	oauth2Repo := repository.NewOAuth2Repository()
	userRepo := repository.NewUserRepository()
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{})
	transactionRepo := repository.NewPayRewardRepository()
	categoryRepo := repository.NewCategoryRepository()

	claimedQuestDomain := NewClaimedQuestDomain(
		claimedQuestRepo,
		questRepo,
		collaboratorRepo,
		followerRepo,
		oauth2Repo,
		userRepo,
		communityRepo,
		transactionRepo,
		categoryRepo,
		&testutil.MockTwitterEndpoint{},
		&testutil.MockDiscordEndpoint{},
		nil, nil,
		&testutil.MockLeaderboard{},
	)

	userDomain := NewUserDomain(
		userRepo, oauth2Repo, followerRepo, nil, communityRepo, nil, nil,
	)

	communityDomain := NewCommunityDomain(communityRepo, collaboratorRepo, userRepo, questRepo, nil, nil, nil)

	newCommunity := entity.Community{
		Base:          entity.Base{ID: uuid.NewString()},
		CreatedBy:     testutil.User1.ID,
		InvitedBy:     sql.NullString{Valid: true, String: testutil.User2.ID},
		InvitedStatus: entity.InvitedStatusUnclaimable,
		Handle:        "new_community",
	}

	err := communityRepo.Create(ctx, &newCommunity)
	require.NoError(t, err)

	// User2 claims invited community reward but community is not enough followers.
	user2Ctx := xcontext.WithRequestUserID(ctx, testutil.User2.ID)
	_, err = claimedQuestDomain.ClaimInvitedCommunity(user2Ctx, &model.ClaimInvitedCommunityRequest{
		WalletAddress: "address",
	})
	require.Error(t, err)
	require.Equal(t, "Not found any claimable invited community", err.Error())

	// User3 follows the community, increase the number of followers by 1.
	// The invited community status is changed to pending.
	user3Ctx := xcontext.WithRequestUserID(ctx, testutil.User3.ID)
	_, err = userDomain.FollowCommunity(user3Ctx, &model.FollowCommunityRequest{CommunityHandle: newCommunity.Handle})
	require.NoError(t, err)

	// Super admin approves the invited community. After that, user2 is eligible
	// for claiming the invited community reward.
	superAdminCtx := xcontext.WithRequestUserID(ctx, testutil.User1.ID)
	_, err = communityDomain.ApproveInvitedCommunities(superAdminCtx, &model.ApproveInvitedCommunitiesRequest{
		CommunityHandles: []string{newCommunity.Handle},
	})
	require.NoError(t, err)

	// User2 reclaims invited community reward and successfully.
	user2Ctx = xcontext.WithRequestUserID(ctx, testutil.User2.ID)
	_, err = claimedQuestDomain.ClaimInvitedCommunity(user2Ctx, &model.ClaimInvitedCommunityRequest{
		WalletAddress: "address",
	})
	require.NoError(t, err)

	// Check transaction in database.
	txs, err := transactionRepo.GetByUserID(ctx, testutil.User2.ID)
	require.NoError(t, err)
	require.Len(t, txs, 1)
	require.Equal(t, testutil.User2.ID, txs[0].UserID)
	require.Equal(t, "Invited community reward of new_community", txs[0].Note)
	require.Equal(t, entity.PayRewardPending, txs[0].Status)
	require.Equal(t, "address", txs[0].Address)
	require.Equal(t, xcontext.Configs(ctx).Quest.InviteCommunityRewardToken, txs[0].Token)
	require.Equal(t, xcontext.Configs(ctx).Quest.InviteCommunityRewardAmount, txs[0].Amount)
}

func Test_fullScenario_Review_Unapprove(t *testing.T) {
	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	collaboratorRepo := repository.NewCollaboratorRepository()
	followerRepo := repository.NewFollowerRepository()
	oauth2Repo := repository.NewOAuth2Repository()
	userRepo := repository.NewUserRepository()
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{})
	payRewardRepo := repository.NewPayRewardRepository()
	categoryRepo := repository.NewCategoryRepository()

	claimedQuestDomain := NewClaimedQuestDomain(
		claimedQuestRepo,
		questRepo,
		collaboratorRepo,
		followerRepo,
		oauth2Repo,
		userRepo,
		communityRepo,
		payRewardRepo,
		categoryRepo,
		&testutil.MockTwitterEndpoint{},
		&testutil.MockDiscordEndpoint{},
		nil, nil,
		&testutil.MockLeaderboard{},
	)

	// TEST CASE 1: Unapprove an accepted claimed-quest.
	ctx = xcontext.WithRequestUserID(ctx, testutil.User1.ID)
	_, err := claimedQuestDomain.Review(ctx, &model.ReviewRequest{
		Action:  string(entity.Pending),
		Comment: "some-comment",
		IDs:     []string{testutil.ClaimedQuest1.ID},
	})
	require.NoError(t, err)

	// Check the new status of claimed-quest.
	claimedQuest, err := claimedQuestRepo.GetByID(ctx, testutil.ClaimedQuest1.ID)
	require.NoError(t, err)
	require.Equal(t, entity.Pending, claimedQuest.Status)
	require.Equal(t, "some-comment", claimedQuest.Comment)

	// Check the points and number of completed quest after unapproving.
	follower, err := followerRepo.Get(ctx, testutil.ClaimedQuest1.UserID, testutil.Community1.ID)
	require.NoError(t, err)
	require.Equal(t, testutil.Follower1.Points-testutil.Quest1.Points, follower.Points)
	require.Equal(t, testutil.Follower1.Quests-1, follower.Quests)

	// TEST CASE 2: Unapprove an rejected claimed-quest.
	_, err = claimedQuestDomain.Review(ctx, &model.ReviewRequest{
		Action:  string(entity.Pending),
		Comment: "some-comment",
		IDs:     []string{testutil.ClaimedQuest2.ID},
	})
	require.NoError(t, err)

	// Check the new status of claimed-quest.
	claimedQuest, err = claimedQuestRepo.GetByID(ctx, testutil.ClaimedQuest2.ID)
	require.NoError(t, err)
	require.Equal(t, entity.Pending, claimedQuest.Status)
	require.Equal(t, "some-comment", claimedQuest.Comment)

	// Check the points and number of completed quest after unapproving (not change).
	follower, err = followerRepo.Get(ctx, testutil.ClaimedQuest2.UserID, testutil.Community1.ID)
	require.NoError(t, err)
	require.Equal(t, testutil.Follower2.Points, follower.Points)
	require.Equal(t, testutil.Follower2.Quests, follower.Quests)

	// TEST CASE 3: Unapprove a pending quest (error).
	_, err = claimedQuestDomain.Review(ctx, &model.ReviewRequest{
		Action:  string(entity.Pending),
		Comment: "some-comment",
		IDs:     []string{testutil.ClaimedQuest3.ID},
	})
	require.Error(t, err)
	require.ErrorIs(t, err, errorx.New(errorx.BadRequest, "Claimed quest must be accepted or rejected"))
}
