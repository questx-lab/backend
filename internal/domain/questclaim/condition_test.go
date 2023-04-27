package questclaim

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
)

func Test_newQuestCondition(t *testing.T) {
	type args struct {
		data map[string]any
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
				data: map[string]any{
					"op":       isCompleted,
					"quest_id": testutil.Quest1.ID,
				},
			},
			want: &questCondition{
				Op:      string(isCompleted),
				QuestID: testutil.Quest1.ID,
			},
			wantErr: nil,
		},
		{
			name: "happy case with is not completed",
			args: args{
				data: map[string]any{
					"op":       isNotCompleted,
					"quest_id": testutil.Quest2.ID,
				},
			},
			want: &questCondition{
				Op:      string(isNotCompleted),
				QuestID: testutil.Quest2.ID,
			},
			wantErr: nil,
		},
		{
			name: "invalid op",
			args: args{
				data: map[string]any{
					"op":       "invalid op",
					"quest_id": testutil.Quest2.ID,
				},
			},
			want:    nil,
			wantErr: errors.New("not found value invalid op in enum questclaim.questConditionOpType"),
		},
		{
			name: "invalid quest id",
			args: args{
				data: map[string]any{
					"op":       isCompleted,
					"quest_id": "invalid quest id",
				},
			},
			want:    nil,
			wantErr: errors.New("record not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testutil.NewMockContext()
			testutil.CreateFixtureDb(ctx)

			got, err := newQuestCondition(
				ctx,
				Factory{
					claimedQuestRepo: repository.NewClaimedQuestRepository(),
					questRepo:        repository.NewQuestRepository(),
				},
				tt.args.data,
				true,
			)

			if tt.wantErr != nil {
				require.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.True(t, reflectutil.PartialEqual(tt.want, got), "%v != %v", tt.want, got)
			}
		})
	}
}

func Test_questCondition_Check(t *testing.T) {
	type fields struct {
		op      string
		questID string
	}

	type args struct {
		ctx xcontext.Context
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
				op:      string(isCompleted),
				questID: testutil.Quest1.ID,
			},
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.ClaimedQuest1.UserID),
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "happy case with is not completed",
			fields: fields{
				op:      string(isCompleted),
				questID: testutil.Quest2.ID,
			},
			args: args{
				ctx: testutil.NewMockContextWithUserID(nil, testutil.ClaimedQuest1.UserID),
			},
			want:    false,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		testutil.CreateFixtureDb(tt.args.ctx)

		t.Run(tt.name, func(t *testing.T) {
			c := &questCondition{
				factory: Factory{claimedQuestRepo: repository.NewClaimedQuestRepository()},
				Op:      tt.fields.op,
				QuestID: tt.fields.questID,
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
		data map[string]any
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
				data: map[string]any{
					"op":   dateBefore,
					"date": "Mar 29 2023",
				},
			},
			want: &dateCondition{
				Op:   string(dateBefore),
				Date: "Mar 29 2023",
			},
			wantErr: nil,
		},
		{
			name: "invalid op",
			args: args{
				data: map[string]any{
					"op":   "invalid op",
					"date": "Mar 29 2023",
				},
			},
			want:    nil,
			wantErr: errors.New("not found value invalid op in enum questclaim.dateConditionOpType"),
		},
		{
			name: "invalid date",
			args: args{
				data: map[string]any{
					"op":   "after",
					"date": "29 Mar 2023",
				},
			},
			want:    nil,
			wantErr: errors.New("parsing time \"29 Mar 2023\" as \"Jan 02 2006\": cannot parse \"29 Mar 2023\" as \"Jan\""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newDateCondition(testutil.NewMockContext(), tt.args.data, true)

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
		op   string
		date string
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
				op:   string(dateBefore),
				date: time.Now().AddDate(0, 0, 1).Format(ConditionDateFormat),
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "happy case with date after",
			fields: fields{
				op:   string(dateAfter),
				date: time.Now().AddDate(0, 0, -1).Format(ConditionDateFormat),
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "failed with date before",
			fields: fields{
				op:   string(dateBefore),
				date: time.Now().AddDate(0, 0, -1).Format(ConditionDateFormat),
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "failed with date after",
			fields: fields{
				op:   string(dateAfter),
				date: time.Now().AddDate(0, 0, 1).Format(ConditionDateFormat),
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "invalid op",
			fields: fields{
				op:   "invalid op",
				date: time.Now().AddDate(0, 0, 1).Format(ConditionDateFormat),
			},
			want:    false,
			wantErr: errors.New("Invalid operator of Date condition"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &dateCondition{
				Op:   tt.fields.op,
				Date: tt.fields.date,
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
