package domain

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"reflect"
	"testing"

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
	ctx := testutil.MockContext(t)
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	followerRepo := repository.NewFollowerRepository()
	followerRoleRepo := repository.NewFollowerRoleRepository()
	userRepo := repository.NewUserRepository(testutil.RedisClient(ctx))
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(ctx))
	categoryRepo := repository.NewCategoryRepository()
	badgeRepo := repository.NewBadgeRepository()
	badgeDetailRepo := repository.NewBadgeDetailRepository()

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
		followerRepo,
		followerRoleRepo,
		userRepo,
		communityRepo,
		categoryRepo,
		badge.NewManager(
			badgeRepo,
			badgeDetailRepo,
			badge.NewRainBowBadgeScanner(badgeRepo, followerRepo),
			badge.NewQuestWarriorBadgeScanner(badgeRepo, followerRepo),
		),
		&testutil.MockLeaderboard{},
		testutil.NewCommunityRoleVerifier(ctx),
		nil,
		testutil.NewQuestFactory(ctx),
		testutil.RedisClient(ctx),
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
	require.Equal(t, "recurrence", err.Error())
}

func Test_claimedQuestDomain_Claim_GivePoint(t *testing.T) {
	ctx := testutil.MockContext(t)
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	followerRepo := repository.NewFollowerRepository()
	followerRoleRepo := repository.NewFollowerRoleRepository()
	userRepo := repository.NewUserRepository(testutil.RedisClient(ctx))
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(ctx))
	badgeRepo := repository.NewBadgeRepository()
	badgeDetailRepo := repository.NewBadgeDetailRepository()
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
		followerRepo,
		followerRoleRepo,
		userRepo,
		communityRepo,
		categoryRepo,
		badge.NewManager(
			badgeRepo,
			badgeDetailRepo,
			badge.NewRainBowBadgeScanner(badgeRepo, followerRepo),
			badge.NewQuestWarriorBadgeScanner(badgeRepo, followerRepo),
		),
		&testutil.MockLeaderboard{},
		testutil.NewCommunityRoleVerifier(ctx),
		nil,
		testutil.NewQuestFactory(ctx),
		testutil.RedisClient(ctx),
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
	require.Equal(t, uint64(1100), follower.Points)

	// Check rainbow (streak) badge.
	myBadge, err := badgeDetailRepo.GetLatest(
		ctx,
		testutil.User1.ID, autoTextQuest.CommunityID.String,
		badge.RainBowBadgeName,
	)
	require.NoError(t, err)
	require.Equal(t, testutil.BadgeRainbow1.ID, myBadge.BadgeID)

	// Check quest warrior badge.
	myBadge, err = badgeDetailRepo.GetLatest(
		ctx,
		testutil.User1.ID,
		autoTextQuest.CommunityID.String,
		badge.QuestWarriorBadgeName,
	)
	require.NoError(t, err)
	require.Equal(t, testutil.BadgeQuestWarrior3.ID, myBadge.BadgeID)
}

func Test_claimedQuestDomain_Claim_ManualText(t *testing.T) {
	ctx := testutil.MockContext(t)
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	followerRepo := repository.NewFollowerRepository()
	followerRoleRepo := repository.NewFollowerRoleRepository()
	userRepo := repository.NewUserRepository(testutil.RedisClient(ctx))
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(ctx))
	categoryRepo := repository.NewCategoryRepository()
	badgeRepo := repository.NewBadgeRepository()
	badgeDetailRepo := repository.NewBadgeDetailRepository()

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
		followerRepo,
		followerRoleRepo,
		userRepo,
		communityRepo,
		categoryRepo,
		badge.NewManager(
			badgeRepo,
			badgeDetailRepo,
			badge.NewRainBowBadgeScanner(badgeRepo, followerRepo),
			badge.NewQuestWarriorBadgeScanner(badgeRepo, followerRepo),
		),
		&testutil.MockLeaderboard{},
		testutil.NewCommunityRoleVerifier(ctx),
		nil,
		testutil.NewQuestFactory(ctx),
		testutil.RedisClient(ctx),
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
	require.Equal(t, "recurrence", err.Error())
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
				ctx: testutil.MockContextWithUserID(t, testutil.User1.ID),
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
				ctx: testutil.MockContextWithUserID(t, testutil.User2.ID),
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
				repository.NewFollowerRepository(),
				repository.NewFollowerRoleRepository(),
				repository.NewUserRepository(testutil.RedisClient(tt.args.ctx)),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(tt.args.ctx)),
				repository.NewCategoryRepository(),
				badge.NewManager(repository.NewBadgeRepository(), repository.NewBadgeDetailRepository()),
				&testutil.MockLeaderboard{},
				testutil.NewCommunityRoleVerifier(tt.args.ctx),
				nil,
				testutil.NewQuestFactory(tt.args.ctx),
				testutil.RedisClient(tt.args.ctx),
			)

			req := httptest.NewRequest("GET", "/claim", nil)
			ctx := xcontext.WithHTTPRequest(tt.args.ctx, req)

			got, err := d.Claim(ctx, tt.args.req)
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
				ctx: testutil.MockContextWithUserID(t, testutil.Community1.CreatedBy),
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
					Rewards:        model.ConvertRewards(testutil.Quest1.Rewards),
					ConditionOp:    string(testutil.Quest1.ConditionOp),
					Conditions:     model.ConvertConditions(testutil.Quest1.Conditions),
				},
				User:           model.ShortUser{ID: testutil.User1.ID},
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
				ctx: testutil.MockContextWithUserID(t, testutil.Community1.CreatedBy),
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
				ctx: testutil.MockContextWithUserID(t, testutil.User2.ID),
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
				userRepo:         repository.NewUserRepository(testutil.RedisClient(tt.args.ctx)),
				categoryRepo:     repository.NewCategoryRepository(),
				communityRepo:    repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(tt.args.ctx)),
				roleVerifier: common.NewCommunityRoleVerifier(
					repository.NewFollowerRoleRepository(),
					repository.NewRoleRepository(),
					repository.NewUserRepository(testutil.RedisClient(tt.args.ctx)),
				),
			}

			req := httptest.NewRequest("GET", "/getClaimedQuests", nil)
			ctx := xcontext.WithHTTPRequest(tt.args.ctx, req)
			got, err := d.Get(ctx, tt.args.req)
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
				ctx: testutil.MockContextWithUserID(t, testutil.Community1.CreatedBy),
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
							Rewards:        model.ConvertRewards(testutil.Quest1.Rewards),
							ConditionOp:    string(testutil.Quest1.ConditionOp),
							Conditions:     model.ConvertConditions(testutil.Quest1.Conditions),
						},
						User:       model.ShortUser{ID: testutil.User1.ID},
						Status:     string(testutil.ClaimedQuest1.Status),
						ReviewerID: testutil.ClaimedQuest1.ReviewerID,
					},
					{
						ID:         testutil.ClaimedQuest2.ID,
						Quest:      model.Quest{ID: testutil.ClaimedQuest2.QuestID},
						User:       model.ShortUser{ID: testutil.ClaimedQuest2.UserID},
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
				ctx: testutil.MockContextWithUserID(t, testutil.Community1.CreatedBy),
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
						User:       model.ShortUser{ID: testutil.ClaimedQuest3.UserID},
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
				ctx: testutil.MockContextWithUserID(t, testutil.Community1.CreatedBy),
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
				ctx: testutil.MockContextWithUserID(t, testutil.Community1.CreatedBy),
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
				ctx: testutil.MockContextWithUserID(t, testutil.User2.ID),
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
				ctx: testutil.MockContextWithUserID(t, testutil.Community1.CreatedBy),
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
						User:       model.ShortUser{ID: testutil.ClaimedQuest1.UserID},
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
				ctx: testutil.MockContextWithUserID(t, testutil.Community1.CreatedBy),
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
						User:       model.ShortUser{ID: testutil.ClaimedQuest2.UserID},
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
				ctx: testutil.MockContextWithUserID(t, testutil.Community1.CreatedBy),
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
						User:       model.ShortUser{ID: testutil.ClaimedQuest3.UserID},
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
				ctx: testutil.MockContextWithUserID(t, testutil.Community1.CreatedBy),
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
						User:       model.ShortUser{ID: testutil.ClaimedQuest3.UserID},
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
				userRepo:         repository.NewUserRepository(testutil.RedisClient(tt.args.ctx)),
				categoryRepo:     repository.NewCategoryRepository(),
				communityRepo:    repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(tt.args.ctx)),
				roleVerifier: common.NewCommunityRoleVerifier(
					repository.NewFollowerRoleRepository(),
					repository.NewRoleRepository(),
					repository.NewUserRepository(testutil.RedisClient(tt.args.ctx)),
				),
			}

			req := httptest.NewRequest("GET", "/getClaimedQuest", nil)
			ctx := xcontext.WithHTTPRequest(tt.args.ctx, req)

			got, err := d.GetList(ctx, tt.args.req)
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
				ctx: testutil.MockContextWithUserID(t, testutil.User3.ID),
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
				ctx: testutil.MockContextWithUserID(t, testutil.User1.ID),
				req: &model.ReviewRequest{
					IDs:    []string{testutil.ClaimedQuest1.ID},
					Action: string(entity.Accepted),
				},
			},
			wantErr: errorx.New(errorx.BadRequest, "Claimed quest claimedQuest1 must be pending"),
		},
		{
			name: "permission denied",
			args: args{
				ctx: testutil.MockContextWithUserID(t, testutil.User2.ID),
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
				repository.NewFollowerRepository(),
				repository.NewFollowerRoleRepository(),
				repository.NewUserRepository(testutil.RedisClient(tt.args.ctx)),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(tt.args.ctx)),
				repository.NewCategoryRepository(),
				badge.NewManager(
					repository.NewBadgeRepository(),
					repository.NewBadgeDetailRepository(),
					badge.NewQuestWarriorBadgeScanner(
						repository.NewBadgeRepository(),
						repository.NewFollowerRepository(),
					),
				),
				&testutil.MockLeaderboard{},
				testutil.NewCommunityRoleVerifier(tt.args.ctx),
				nil,
				testutil.NewQuestFactory(tt.args.ctx),
				testutil.RedisClient(tt.args.ctx),
			)
			req := httptest.NewRequest("GET", "/review", nil)
			ctx := xcontext.WithHTTPRequest(tt.args.ctx, req)

			got, err := d.Review(ctx, tt.args.req)
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
				ctx: testutil.MockContextWithUserID(t, testutil.User3.ID),
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
				ctx: testutil.MockContextWithUserID(t, testutil.User3.ID),
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
				ctx: testutil.MockContextWithUserID(t, testutil.User1.ID),
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
				ctx: testutil.MockContextWithUserID(t, testutil.User1.ID),
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
				ctx: testutil.MockContextWithUserID(t, testutil.User2.ID),
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
				ctx: testutil.MockContextWithUserID(t, testutil.User1.ID),
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
			req := httptest.NewRequest("GET", "/reviewAll", nil)
			ctx := xcontext.WithHTTPRequest(tt.args.ctx, req)
			err := claimedQuestRepo.Create(ctx, &entity.ClaimedQuest{
				Base:    entity.Base{ID: "claimed_quest_test_1"},
				QuestID: testutil.Quest1.ID,
				UserID:  testutil.User2.ID,
				Status:  entity.Pending,
			})
			require.NoError(t, err)

			err = claimedQuestRepo.Create(ctx, &entity.ClaimedQuest{
				Base:    entity.Base{ID: "claimed_quest_test_2"},
				QuestID: testutil.Quest1.ID,
				UserID:  testutil.User3.ID,
				Status:  entity.Pending,
			})
			require.NoError(t, err)

			err = claimedQuestRepo.Create(ctx, &entity.ClaimedQuest{
				Base:    entity.Base{ID: "claimed_quest_test_3"},
				QuestID: testutil.Quest2.ID,
				UserID:  testutil.User1.ID,
				Status:  entity.Pending,
			})
			require.NoError(t, err)

			err = claimedQuestRepo.Create(ctx, &entity.ClaimedQuest{
				Base:    entity.Base{ID: "claimed_quest_test_4"},
				QuestID: testutil.Quest3.ID,
				UserID:  testutil.User1.ID,
				Status:  entity.Accepted,
			})
			require.NoError(t, err)

			d := NewClaimedQuestDomain(
				repository.NewClaimedQuestRepository(),
				repository.NewQuestRepository(&testutil.MockSearchCaller{}),
				repository.NewFollowerRepository(),
				repository.NewFollowerRoleRepository(),
				repository.NewUserRepository(testutil.RedisClient(ctx)),
				repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(ctx)),
				repository.NewCategoryRepository(),
				badge.NewManager(
					repository.NewBadgeRepository(),
					repository.NewBadgeDetailRepository(),
					badge.NewQuestWarriorBadgeScanner(
						repository.NewBadgeRepository(),
						repository.NewFollowerRepository(),
					),
				),
				&testutil.MockLeaderboard{},
				testutil.NewCommunityRoleVerifier(ctx),
				nil,
				testutil.NewQuestFactory(ctx),
				testutil.RedisClient(ctx),
			)
			got, err := d.ReviewAll(ctx, tt.args.req)
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

func Test_fullScenario_Review_Unapprove(t *testing.T) {
	ctx := testutil.MockContext(t)
	testutil.CreateFixtureDb(ctx)
	claimedQuestRepo := repository.NewClaimedQuestRepository()
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	followerRepo := repository.NewFollowerRepository()
	followerRoleRepo := repository.NewFollowerRoleRepository()
	userRepo := repository.NewUserRepository(testutil.RedisClient(ctx))
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{}, testutil.RedisClient(ctx))
	categoryRepo := repository.NewCategoryRepository()

	claimedQuestDomain := NewClaimedQuestDomain(
		claimedQuestRepo,
		questRepo,
		followerRepo,
		followerRoleRepo,
		userRepo,
		communityRepo,
		categoryRepo, nil,
		&testutil.MockLeaderboard{},
		testutil.NewCommunityRoleVerifier(ctx),
		nil,
		testutil.NewQuestFactory(ctx),
		testutil.RedisClient(ctx),
	)

	// TEST CASE 1: Unapprove an accepted claimed-quest.
	ctx = xcontext.WithRequestUserID(ctx, testutil.User1.ID)
	req := httptest.NewRequest("GET", "/review", nil)
	ctx = xcontext.WithHTTPRequest(ctx, req)

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
	require.ErrorIs(t, err, errorx.New(errorx.BadRequest, "Claimed quest claimedQuest3 must be accepted or rejected"))
}
