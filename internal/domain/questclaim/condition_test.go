package questclaim

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func Test_newQuestCondition(t *testing.T) {
	db := testutil.CreateFixtureDb()
	claimedQuestRepo := repository.NewClaimedQuestRepository(db)
	questRepo := repository.NewQuestRepository(db)

	type args struct {
		condition entity.Condition
	}

	tests := []struct {
		name    string
		args    args
		want    *questCondition
		wantErr error
	}{
		{
			name: "happy case with is completed",
			args: args{
				condition: entity.Condition{
					Op:    string(isCompleted),
					Value: testutil.Quest1.ID,
				}},
			want: &questCondition{
				claimedQuestRepo: claimedQuestRepo,
				op:               isCompleted,
				questID:          testutil.Quest1.ID,
			},
			wantErr: nil,
		},
		{
			name: "happy case with is not completed",
			args: args{
				condition: entity.Condition{
					Op:    string(isNotCompleted),
					Value: testutil.Quest2.ID,
				}},
			want: &questCondition{
				claimedQuestRepo: claimedQuestRepo,
				op:               isNotCompleted,
				questID:          testutil.Quest2.ID,
			},
			wantErr: nil,
		},
		{
			name: "invalid op",
			args: args{
				condition: entity.Condition{
					Op:    "invalid op",
					Value: testutil.Quest2.ID,
				}},
			want:    nil,
			wantErr: errors.New("not found value invalid op in enum questclaim.questConditionOpType"),
		},
		{
			name: "invalid quest id",
			args: args{
				condition: entity.Condition{
					Op:    string(isCompleted),
					Value: "invalid quest id",
				}},
			want:    nil,
			wantErr: errors.New("record not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newQuestCondition(
				testutil.NewMockContext(),
				tt.args.condition,
				claimedQuestRepo,
				questRepo,
			)

			if tt.wantErr != nil {
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

func Test_questCondition_Check(t *testing.T) {
	db := testutil.CreateFixtureDb()
	claimedQuestRepo := repository.NewClaimedQuestRepository(db)

	type fields struct {
		op      questConditionOpType
		questID string
	}

	type args struct {
		ctx router.Context
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name: "happy case with is completed",
			fields: fields{
				op:      isCompleted,
				questID: testutil.Quest1.ID,
			},
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.ClaimedQuest1.UserID),
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "happy case with is not completed",
			fields: fields{
				op:      isCompleted,
				questID: testutil.Quest2.ID,
			},
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.ClaimedQuest1.UserID),
			},
			want:    false,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &questCondition{
				claimedQuestRepo: claimedQuestRepo,
				op:               tt.fields.op,
				questID:          tt.fields.questID,
			}

			got, err := c.Check(tt.args.ctx)
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

func Test_newDateCondition(t *testing.T) {
	type args struct {
		condition entity.Condition
	}
	tests := []struct {
		name    string
		args    args
		want    *dateCondition
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				condition: entity.Condition{
					Op:    "before",
					Value: "Mar 29 2023",
				},
			},
			want: &dateCondition{
				op:   dateBefore,
				date: time.Date(2023, time.March, 29, 0, 0, 0, 0, time.UTC),
			},
			wantErr: nil,
		},
		{
			name: "invalid op",
			args: args{
				condition: entity.Condition{
					Op:    "invalid op",
					Value: "Mar 29 2023",
				},
			},
			want:    nil,
			wantErr: errors.New("not found value invalid op in enum questclaim.dateConditionOpType"),
		},
		{
			name: "invalid date",
			args: args{
				condition: entity.Condition{
					Op:    "after",
					Value: "29 Mar 2023",
				},
			},
			want:    nil,
			wantErr: errors.New("parsing time \"29 Mar 2023\" as \"Jan 02 2006\": cannot parse \"29 Mar 2023\" as \"Jan\""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newDateCondition(testutil.NewMockContext(), tt.args.condition)

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

func Test_dateCondition_Check(t *testing.T) {
	type fields struct {
		op   dateConditionOpType
		date time.Time
	}

	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr error
	}{
		{
			name: "happy case with date before",
			fields: fields{
				op:   dateBefore,
				date: time.Now().AddDate(0, 0, 1),
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "happy case with date after",
			fields: fields{
				op:   dateAfter,
				date: time.Now().AddDate(0, 0, -1),
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "failed with date before",
			fields: fields{
				op:   dateBefore,
				date: time.Now().AddDate(0, 0, -1),
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "failed with date after",
			fields: fields{
				op:   dateAfter,
				date: time.Now().AddDate(0, 0, 1),
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "invalid op",
			fields: fields{
				op:   "invalid op",
				date: time.Now().AddDate(0, 0, 1),
			},
			want:    false,
			wantErr: errors.New("Invalid operator of Date condition"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &dateCondition{
				op:   tt.fields.op,
				date: tt.fields.date,
			}

			got, err := c.Check(testutil.NewMockContext())
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
