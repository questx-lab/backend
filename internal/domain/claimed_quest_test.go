package domain

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func Test_claimedQuestDomain_Claim(t *testing.T) {
	db := testutil.CreateFixtureDb()
	claimedQuestRepo := repository.NewClaimedQuestRepository(db)
	questRepo := repository.NewQuestRepository(db)

	autoTextQuest := &entity.Quest{
		Base: entity.Base{
			ID: "auto text quest",
		},
		ProjectID:      testutil.Project1.ID,
		Type:           entity.Text,
		Status:         entity.Published,
		CategoryIDs:    []string{},
		Recurrence:     entity.Daily,
		ValidationData: `{"auto_validate": true, "answer": "Foo"}`,
		ConditionOp:    entity.Or,
		Conditions: []entity.Condition{{
			Type:  entity.QuestCondition,
			Op:    "is completed",
			Value: testutil.Quest1.ID,
		}},
	}

	manualTextQuest := &entity.Quest{
		Base: entity.Base{
			ID: "manual text quest",
		},
		ProjectID:      testutil.Project1.ID,
		Type:           entity.Text,
		Status:         entity.Published,
		CategoryIDs:    []string{},
		Recurrence:     entity.Daily,
		ValidationData: `{"auto_validate": false}`,
		ConditionOp:    entity.Or,
	}

	err := questRepo.Create(context.Background(), autoTextQuest)
	require.NoError(t, err)

	err = questRepo.Create(context.Background(), manualTextQuest)
	require.NoError(t, err)

	type args struct {
		ctx router.Context
		req *model.ClaimQuestRequest
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "rejected with wrong answer",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.ClaimQuestRequest{
					QuestID: autoTextQuest.ID,
					Input:   "Bar",
				},
			},
			want:    string(entity.Rejected),
			wantErr: false,
		},
		{
			name: "happy case with auto review",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.ClaimQuestRequest{
					QuestID: autoTextQuest.ID,
					Input:   "Foo",
				},
			},
			want:    string(entity.AutoAccepted),
			wantErr: false,
		},
		{
			name: "failed to claim autoTextQuest again (daily recurrence)",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.ClaimQuestRequest{
					QuestID: autoTextQuest.ID,
					Input:   "Foo",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "happy case with manual review",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.ClaimQuestRequest{
					QuestID: manualTextQuest.ID,
					Input:   "any",
				},
			},
			want:    string(entity.Pending),
			wantErr: false,
		},
		{
			name: "failed to claim manualTextQuest again (daily recurrence)",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.ClaimQuestRequest{
					QuestID: manualTextQuest.ID,
					Input:   "any",
				},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &claimedQuestDomain{
				claimedQuestRepo: claimedQuestRepo,
				questRepo:        questRepo,
			}

			got, err := d.Claim(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("claimedQuestDomain.Claim() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != "" {
				require.Equal(t, tt.want, got.Status)
			}
		})
	}
}

func Test_claimedQuestDomain_Get(t *testing.T) {
	db := testutil.CreateFixtureDb()
	claimedQuestRepo := repository.NewClaimedQuestRepository(db)
	questRepo := repository.NewQuestRepository(db)

	type args struct {
		ctx router.Context
		req *model.GetClaimedQuestRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.GetClaimedQuestResponse
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
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
			wantErr: false,
		},
		{
			name: "invalid id",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.GetClaimedQuestRequest{
					ID: "invalid id",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &claimedQuestDomain{
				claimedQuestRepo: claimedQuestRepo,
				questRepo:        questRepo,
			}

			got, err := d.Get(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("claimedQuestDomain.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("claimedQuestDomain.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_claimedQuestDomain_GetList(t *testing.T) {
	db := testutil.CreateFixtureDb()
	claimedQuestRepo := repository.NewClaimedQuestRepository(db)
	questRepo := repository.NewQuestRepository(db)

	type args struct {
		ctx router.Context
		req *model.GetListClaimedQuestRequest
	}

	tests := []struct {
		name    string
		args    args
		want    *model.GetListClaimedQuestResponse
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    0,
					Limit:     2,
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						QuestID:    testutil.ClaimedQuest1.QuestID,
						UserID:     testutil.ClaimedQuest1.UserID,
						Status:     string(testutil.ClaimedQuest1.Status),
						ReviewerID: testutil.ClaimedQuest1.ReviewerID,
						ReviewerAt: testutil.ClaimedQuest1.ReviewerAt.Format(time.RFC3339Nano),
					},
					{
						QuestID:    testutil.ClaimedQuest2.QuestID,
						UserID:     testutil.ClaimedQuest2.UserID,
						Status:     string(testutil.ClaimedQuest2.Status),
						ReviewerID: testutil.ClaimedQuest2.ReviewerID,
						ReviewerAt: testutil.ClaimedQuest2.ReviewerAt.Format(time.RFC3339Nano),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "happy case with custom offset",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    2,
					Limit:     1,
				},
			},
			want: &model.GetListClaimedQuestResponse{
				ClaimedQuests: []model.ClaimedQuest{
					{
						QuestID:    testutil.ClaimedQuest3.QuestID,
						UserID:     testutil.ClaimedQuest3.UserID,
						Status:     string(testutil.ClaimedQuest3.Status),
						ReviewerID: testutil.ClaimedQuest3.ReviewerID,
						ReviewerAt: testutil.ClaimedQuest3.ReviewerAt.Format(time.RFC3339Nano),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nagative limit",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    2,
					Limit:     -1,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "exceed the maximum limit",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    2,
					Limit:     51,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &claimedQuestDomain{
				claimedQuestRepo: claimedQuestRepo,
				questRepo:        questRepo,
			}

			got, err := d.GetList(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("claimedQuestDomain.GetList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("claimedQuestDomain.GetList() = %v, want %v", got, tt.want)
			}
		})
	}
}
