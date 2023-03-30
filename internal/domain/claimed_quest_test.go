package domain

import (
	"context"
	"errors"
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

func Test_claimedQuestDomain_Claim_AutoText(t *testing.T) {
	db := testutil.CreateFixtureDb()
	claimedQuestRepo := repository.NewClaimedQuestRepository(db)
	questRepo := repository.NewQuestRepository(db)
	collaboratorRepo := repository.NewCollaboratorRepository(db)

	autoTextQuest := &entity.Quest{
		Base: entity.Base{
			ID: "auto text quest",
		},
		ProjectID:      testutil.Project1.ID,
		Type:           entity.Text,
		Status:         entity.QuestActive,
		CategoryIDs:    []string{},
		Recurrence:     entity.Daily,
		ValidationData: `{"auto_validate": true, "answer": "Foo"}`,
		ConditionOp:    entity.Or,
	}

	err := questRepo.Create(context.Background(), autoTextQuest)
	require.NoError(t, err)

	d := NewClaimedQuestDomain(claimedQuestRepo, questRepo, collaboratorRepo)

	steps := []struct {
		name       string
		userID     string
		input      string
		wantError  error
		wantStatus string
	}{
		{
			name:       "user1 cannot claim quest with a wrong answer",
			userID:     testutil.User1.ID,
			input:      "wrong answer",
			wantError:  nil,
			wantStatus: "auto_rejected",
		},
		{
			name:       "user1 claims quest again but with a correct answer",
			userID:     testutil.User1.ID,
			input:      "Foo",
			wantError:  nil,
			wantStatus: "auto_accepted",
		},
		{
			name:      "user1 cannot claims quest again because the daily recurrence",
			userID:    testutil.User1.ID,
			input:     "Foo",
			wantError: errors.New("This quest cannot be claimed now"),
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			ctx := testutil.NewMockContextWithUserID(step.userID)
			got, err := d.Claim(ctx, &model.ClaimQuestRequest{
				QuestID: autoTextQuest.ID,
				Input:   step.input,
			})

			if step.wantError == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, step.wantError.Error(), err.Error())
			}

			if step.wantStatus != "" {
				require.Equal(t, step.wantStatus, got.Status)
			}
		})
	}
}

func Test_claimedQuestDomain_Claim_ManualText(t *testing.T) {
	db := testutil.CreateFixtureDb()
	claimedQuestRepo := repository.NewClaimedQuestRepository(db)
	questRepo := repository.NewQuestRepository(db)
	collaboratorRepo := repository.NewCollaboratorRepository(db)

	autoTextQuest := &entity.Quest{
		Base: entity.Base{
			ID: "manual text quest",
		},
		ProjectID:      testutil.Project1.ID,
		Type:           entity.Text,
		Status:         entity.QuestActive,
		CategoryIDs:    []string{},
		Recurrence:     entity.Daily,
		ValidationData: `{"auto_validate": false}`,
		ConditionOp:    entity.Or,
	}

	err := questRepo.Create(context.Background(), autoTextQuest)
	require.NoError(t, err)

	d := NewClaimedQuestDomain(claimedQuestRepo, questRepo, collaboratorRepo)

	steps := []struct {
		name       string
		wantError  error
		wantStatus string
	}{
		{
			name:       "need to wait for a manual review if claim manual text quest",
			wantError:  nil,
			wantStatus: "pending",
		},
		{
			name:      "cannot claim the quest again while the quest is pending",
			wantError: errors.New("This quest cannot be claimed now"),
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			ctx := testutil.NewMockContextWithUserID(testutil.User1.ID)
			got, err := d.Claim(ctx, &model.ClaimQuestRequest{
				QuestID: autoTextQuest.ID,
				Input:   "any anwser",
			})

			if step.wantError == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, step.wantError.Error(), err.Error())
			}

			if step.wantStatus != "" {
				require.Equal(t, step.wantStatus, got.Status)
			}
		})
	}
}

func Test_claimedQuestDomain_Claim(t *testing.T) {

	type args struct {
		ctx router.Context
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
				ctx: testutil.NewMockContextWithUserID(testutil.User1.ID),
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
				ctx: testutil.NewMockContextWithUserID(testutil.User2.ID),
				req: &model.ClaimQuestRequest{
					QuestID: testutil.Quest2.ID,
				},
			},
			wantErr: errors.New("This quest cannot be claimed now"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.CreateFixtureDb()
			claimedQuestRepo := repository.NewClaimedQuestRepository(db)
			questRepo := repository.NewQuestRepository(db)
			collaboratorRepo := repository.NewCollaboratorRepository(db)

			d := &claimedQuestDomain{
				claimedQuestRepo: claimedQuestRepo,
				questRepo:        questRepo,
				collaboratorRepo: collaboratorRepo,
			}

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
		ctx router.Context
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
				ctx: testutil.NewMockContextWithUserID(testutil.Collaborator1.UserID),
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
				ctx: testutil.NewMockContextWithUserID(testutil.Collaborator1.UserID),
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
				ctx: testutil.NewMockContextWithUserID(testutil.User2.ID),
				req: &model.GetClaimedQuestRequest{
					ID: testutil.ClaimedQuest1.ID,
				},
			},
			want:    nil,
			wantErr: errors.New("User does not have permission"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.CreateFixtureDb()
			claimedQuestRepo := repository.NewClaimedQuestRepository(db)
			questRepo := repository.NewQuestRepository(db)
			collaboratorRepo := repository.NewCollaboratorRepository(db)

			d := &claimedQuestDomain{
				claimedQuestRepo: claimedQuestRepo,
				questRepo:        questRepo,
				collaboratorRepo: collaboratorRepo,
			}

			got, err := d.Get(tt.args.ctx, tt.args.req)
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

func Test_claimedQuestDomain_GetList(t *testing.T) {
	type args struct {
		ctx router.Context
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
				ctx: testutil.NewMockContextWithUserID(testutil.Collaborator1.UserID),
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
			wantErr: nil,
		},
		{
			name: "happy case with custom offset",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.Collaborator1.UserID),
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
			wantErr: nil,
		},
		{
			name: "nagative limit",
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.Collaborator1.UserID),
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
				ctx: testutil.NewMockContextWithUserID(testutil.Collaborator1.UserID),
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
				ctx: testutil.NewMockContextWithUserID(testutil.User2.ID),
				req: &model.GetListClaimedQuestRequest{
					ProjectID: testutil.Project1.ID,
					Offset:    2,
					Limit:     51,
				},
			},
			want:    nil,
			wantErr: errors.New("User does not have permission"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.CreateFixtureDb()
			claimedQuestRepo := repository.NewClaimedQuestRepository(db)
			questRepo := repository.NewQuestRepository(db)
			collaboratorRepo := repository.NewCollaboratorRepository(db)

			d := &claimedQuestDomain{
				claimedQuestRepo: claimedQuestRepo,
				questRepo:        questRepo,
				collaboratorRepo: collaboratorRepo,
			}

			got, err := d.GetList(tt.args.ctx, tt.args.req)
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
